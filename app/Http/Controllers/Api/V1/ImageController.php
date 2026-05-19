<?php

namespace App\Http\Controllers\Api\V1;

use App\Exceptions\UploadException;
use App\Http\Controllers\Controller;
use App\Models\Image;
use App\Models\User;
use App\Services\ImageService;
use App\Services\UserService;
use App\Utils;
use Illuminate\Auth\AuthenticationException;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Auth;

class ImageController extends Controller
{
    /**
     * @throws AuthenticationException
     */
    public function upload(Request $request, ImageService $service): Response
    {
        if ($request->hasHeader('Authorization')) {
            $guards = array_keys(config('auth.guards'));

            if (empty($guards)) {
                $guards = [null];
            }

            foreach ($guards as $guard) {
                if (Auth::guard($guard)->check()) {
                    Auth::shouldUse($guard);
                    break;
                }
            }

            if (! Auth::check()) {
                throw new AuthenticationException('Authentication failed.');
            }
        }

        try {
            $image = $service->store($request);
        } catch (UploadException $e) {
            return $this->fail($e->getMessage());
        } catch (\Throwable $e) {
            Utils::e($e, 'Api 上传文件时发生异常');
            if (config('app.debug')) {
                return $this->fail($e->getMessage());
            }
            return $this->fail('服务异常，请稍后再试');
        }
        return $this->success('上传成功', $image->setAppends(['pathname', 'links'])->only(
            'key', 'name', 'pathname', 'origin_name', 'size', 'mimetype', 'extension', 'md5', 'sha1', 'links'
        ));
    }

    public function images(Request $request): Response
    {
        /** @var User $user */
        $user = Auth::user();

        $images = $user->images()->filter($request)->paginate(40)->withQueryString();
        $images->getCollection()->each(function (Image $image) {
            $image->human_date = $image->created_at->diffForHumans();
            $image->date = $image->created_at->format('Y-m-d H:i:s');
            $image->append(['pathname', 'links'])->setVisible([
                'album', 'key', 'name', 'pathname', 'origin_name', 'size', 'mimetype', 'extension', 'md5', 'sha1',
                'width', 'height', 'links', 'human_date', 'date',
            ]);
        });
        return $this->success('success', $images);
    }

    public function destroy(Request $request): Response
    {
        /** @var User $user */
        $user = Auth::user();
        (new UserService())->deleteImages([$request->route('key')], $user, 'key');
        return $this->success('删除成功');
    }

    public function random(Request $request): Response
    {
        /** @var User $user */
        $user = Auth::user();

        $query = $user->images();

        // 相册筛选
        if ($request->has('album_id')) {
            $albumId = (int) $request->query('album_id');
            $albumId === 0
                ? $query->whereNull('album_id')
                : $query->where('album_id', $albumId);
        }

        // 尺寸范围
        if ($minW = $request->query('min_width')) {
            $query->where('width', '>=', (int) $minW);
        }
        if ($maxW = $request->query('max_width')) {
            $query->where('width', '<=', (int) $maxW);
        }
        if ($minH = $request->query('min_height')) {
            $query->where('height', '>=', (int) $minH);
        }
        if ($maxH = $request->query('max_height')) {
            $query->where('height', '<=', (int) $maxH);
        }

        // 画幅方向
        if ($orientation = $request->query('orientation')) {
            match ($orientation) {
                'landscape' => $query->whereRaw('width > height'),
                'portrait'  => $query->whereRaw('width < height'),
                'square'    => $query->whereRaw('width = height'),
                default     => null,
            };
        }

        // 画幅比例 (如 16:9, 4:3, 或直接传小数 1.778)
        if ($ratio = $request->query('ratio')) {
            $tolerance = 0.05; // ±5%
            if (str_contains($ratio, ':')) {
                [$w, $h] = explode(':', $ratio);
                $ratio = (float) $w / (float) $h;
            }
            $ratio = (float) $ratio;
            $lo = $ratio * (1 - $tolerance);
            $hi = $ratio * (1 + $tolerance);
            // SQLite PDO 参数化浮点有坑，直接内联数值
            $query->whereRaw("width * 1.0 / height BETWEEN {$lo} AND {$hi}");
        }

        return $this->returnImage($query);
    }

    /**
     * 自适应设备：根据 User-Agent 自动返回适配当前屏幕的图片。
     *
     * 设备 → 画幅映射：
     *   iPhone / Android 手机 → 竖版 portrait
     *   iPad               → 横版 4:3
     *   Android 平板        → 横版 16:9
     *   Windows / Mac       → 横版 16:9
     *   其他                → 横版 (fallback)
     *
     * 渐进回退：精确匹配失败 → 只按方向 → 完全随机
     */
    public function adaptive(Request $request): Response
    {
        /** @var User $user */
        $user = Auth::user();

        $ua = $request->header('User-Agent', '');

        // 设备检测
        $isIPhone = str_contains($ua, 'iPhone') || str_contains($ua, 'iPod');
        $isIPad = str_contains($ua, 'iPad');
        $isAndroid = str_contains($ua, 'Android');
        $isAndroidMobile = $isAndroid && str_contains($ua, 'Mobile');
        $isDesktop = !$isIPhone && !$isIPad && !$isAndroid;

        // 决定目标参数
        if ($isIPhone) {
            $orientation = 'portrait';
            $ratio = null;       // iPhone 比例多样 (19.5:9, 16:9...)，只用方向
        } elseif ($isIPad) {
            $orientation = 'landscape';
            $ratio = '4:3';
        } elseif ($isAndroidMobile) {
            $orientation = 'portrait';
            $ratio = null;
        } elseif ($isAndroid) {
            // Android 平板
            $orientation = 'landscape';
            $ratio = '16:9';
        } else {
            // 桌面 / 其他
            $orientation = 'landscape';
            $ratio = '16:9';
        }

        $baseQuery = $user->images();

        // --- 第一轮：精确匹配 (orientation + ratio) ---
        if ($ratio) {
            $query1 = clone $baseQuery;
            $this->applyOrientation($query1, $orientation);
            $this->applyRatio($query1, $ratio);
            $image = $query1->inRandomOrder()->first();
            if ($image) {
                return $this->returnImage($query1, $image);
            }
        }

        // --- 第二轮：只按方向 ---
        $query2 = clone $baseQuery;
        $this->applyOrientation($query2, $orientation);
        $image = $query2->inRandomOrder()->first();
        if ($image) {
            return $this->returnImage($query2, $image);
        }

        // --- 第三轮：完全随机 ---
        return $this->returnImage($baseQuery);
    }

    private function applyOrientation($query, string $orientation): void
    {
        match ($orientation) {
            'landscape' => $query->whereRaw('width > height'),
            'portrait'  => $query->whereRaw('width < height'),
            'square'    => $query->whereRaw('width = height'),
            default     => null,
        };
    }

    private function applyRatio($query, string $ratio): void
    {
        $tolerance = 0.05;
        if (str_contains($ratio, ':')) {
            [$w, $h] = explode(':', $ratio);
            $ratio = (float) $w / (float) $h;
        }
        $ratio = (float) $ratio;
        $lo = $ratio * (1 - $tolerance);
        $hi = $ratio * (1 + $tolerance);
        $query->whereRaw("width * 1.0 / height BETWEEN {$lo} AND {$hi}");
    }

    private function returnImage($query, $image = null): Response
    {
        $image ??= $query->inRandomOrder()->first();

        if (! $image) {
            return $this->fail('没有符合条件的图片');
        }

        $image->human_date = $image->created_at->diffForHumans();
        $image->date = $image->created_at->format('Y-m-d H:i:s');
        $image->append(['pathname', 'links'])->setVisible([
            'album', 'key', 'name', 'pathname', 'origin_name', 'size', 'mimetype', 'extension', 'md5', 'sha1',
            'width', 'height', 'links', 'human_date', 'date',
        ]);

        return $this->success('success', $image);
    }
}
