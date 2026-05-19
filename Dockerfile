FROM php:8.2-fpm

# Use Tencent Cloud Debian mirror
RUN echo 'deb https://mirrors.cloud.tencent.com/debian/ trixie main' > /etc/apt/sources.list && \
    echo 'deb https://mirrors.cloud.tencent.com/debian/ trixie-updates main' >> /etc/apt/sources.list && \
    echo 'deb https://mirrors.cloud.tencent.com/debian-security/ trixie-security main' >> /etc/apt/sources.list && \
    apt-get update -qq && \
    apt-get install -y -qq libzip-dev libmagickwand-dev && \
    docker-php-ext-install pdo pdo_sqlite bcmath exif gd zip && \
    pecl install imagick && docker-php-ext-enable imagick && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /var/www/html
