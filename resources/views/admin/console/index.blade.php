@section('title', '系统控制台')

<x-app-layout>
    @if(config('app.debug'))
        <p class="mt-4 p-2 rounded-md text-sm bg-red-500 text-white">
            <i class="fas fa-exclamation-triangle"></i>
            当前系统 debug 已被打开，敏感信息暴露在外，可能会被利用从而影响系统稳定性，生产环境中请务必关闭！
        </p>
    @endif
    <div class="my-6 md:my-9">
        <p class="mb-3 font-semibold text-lg text-gray-700">概览</p>
        <div class="relative grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-violet-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-violet-400 truncate">
                        {{ \App\Utils::shortenNumber(\App\Models\Image::query()->count()) }}
                    </p>
                    <p class="text-sm text-gray-400">图片数量</p>
                </div>
                <i class="fas fa-images text-violet-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-emerald-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-emerald-400 truncate">
                        {{ \App\Utils::shortenNumber(\App\Models\Album::query()->count()) }}
                    </p>
                    <p class="text-sm text-gray-400">相册数量</p>
                </div>
                <i class="fas fa-tags text-emerald-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-sky-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-sky-400 truncate">
                        {{ \App\Utils::shortenNumber(\App\Models\User::query()->count()) }}
                    </p>
                    <p class="text-sm text-gray-400">用户数量</p>
                </div>
                <i class="fas fa-users text-sky-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-cyan-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-cyan-400 truncate">
                        {{ \App\Utils::formatSize(\App\Models\Image::query()->sum('size') * 1024) }}
                    </p>
                    <p class="text-sm text-gray-400">占用储存</p>
                </div>
                <i class="fas fa-server text-cyan-500/50 text-2xl"></i>
            </div>

            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-gray-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-gray-300 truncate">{{ \App\Utils::shortenNumber($numbers['today']) }}</p>
                    <p class="text-sm text-gray-400">今日上传</p>
                </div>
                <i class="fas fa-upload text-gray-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-gray-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-gray-300 truncate">{{ \App\Utils::shortenNumber($numbers['yesterday']) }}</p>
                    <p class="text-sm text-gray-400">昨日上传</p>
                </div>
                <i class="fas fa-upload text-gray-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-gray-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-gray-300 truncate">{{ \App\Utils::shortenNumber($numbers['week']) }}</p>
                    <p class="text-sm text-gray-400">本周上传</p>
                </div>
                <i class="fas fa-upload text-gray-500/50 text-2xl"></i>
            </div>
            <div class="flex justify-between rounded-xl bg-surface border border-white/5 p-4 overflow-hidden hover:border-gray-500/30 transition-all duration-200">
                <div class="flex flex-col justify-between space-y-2 w-[80%]">
                    <p class="font-bold text-2xl text-gray-300 truncate">{{ \App\Utils::shortenNumber($numbers['month']) }}</p>
                    <p class="text-sm text-gray-400">本月上传</p>
                </div>
                <i class="fas fa-upload text-gray-500/50 text-2xl"></i>
            </div>
        </div>

        <p class="mb-3 font-semibold text-lg text-gray-300">趋势</p>
        <div class="relative p-4 rounded-xl bg-surface border border-white/5 h-80 mb-8" id="chart">
            <canvas></canvas>
        </div>

        <p class="mb-3 font-semibold text-lg text-gray-300">系统情况</p>
        <div class="relative rounded-xl bg-surface border border-white/5 mb-8 overflow-hidden">
            <dl>
                <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">操作系统</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {{ php_uname() }}
                    </dd>
                </div>
                <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">运行环境</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {{ request()->server('SERVER_SOFTWARE') }}
                    </dd>
                </div>
                <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">PHP 版本</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {{ phpversion() }}
                    </dd>
                </div>
                <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">文件上传限制</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {{ ini_get("upload_max_filesize") }}
                    </dd>
                </div>
                <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">POST 数据最大限制</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {{ ini_get('post_max_size') }}
                    </dd>
                </div>
            </dl>
        </div>

        <p class="mb-3 font-semibold text-lg text-gray-300">软件信息</p>
        <div class="relative rounded-xl bg-surface border border-white/5 mb-8 overflow-hidden">
            <dl>
                <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">软件版本</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ \App\Utils::config(\App\Enums\ConfigKey::AppVersion) }}</dd>
                </div>
                <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">官方网站</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        <a target="_blank" class="hover:text-blue-500" href="https://www.lsky.pro">https://www.lsky.pro</a>
                    </dd>
                </div>
                <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">使用手册</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        <a target="_blank" class="hover:text-blue-500" href="https://docs.lsky.pro">https://docs.lsky.pro</a>
                    </dd>
                </div>
                <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt class="text-sm font-medium text-gray-500">仓库地址</dt>
                    <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        <a target="_blank" class="hover:text-blue-500" href="https://github.com/lsky-org/lsky-pro">https://github.com/lsky-org/lsky-pro</a>
                    </dd>
                </div>
            </dl>
        </div>
    </div>

    @push('scripts')
        <script src="{{ asset('js/echarts/echarts.min.js') }}"></script>
        <script>
            $(function () {
                'use strict'
                let chartDom = document.getElementById('chart');
                let myChart = echarts.init(chartDom);
                let options;

                options = {
                    responsive: true,
                    title: {
                        text: '近 30 天内统计'
                    },
                    tooltip: {
                        trigger: 'axis'
                    },
                    legend: {
                        top: '10%',
                        type: 'scroll',
                        data: @json($fields)
                    },
                    grid: {
                        left: '3%',
                        right: '3%',
                        bottom: '3%',
                        containLabel: true
                    },
                    toolbox: {
                        show: true,
                        feature: {
                            magicType: {
                                type: ["line", "bar"]
                            },
                            saveAsImage: {}
                        }
                    },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: @json($dates)
                    },
                    yAxis: {
                        type: 'value',
                        minInterval: 1,
                    },
                    series: @json($datasets)
                };

                options && myChart.setOption(options);

                window.onresize = function() {
                    myChart.resize();
                }
            })
        </script>
    @endpush

</x-app-layout>
