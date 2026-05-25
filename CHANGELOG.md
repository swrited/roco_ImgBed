# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of Sino-ImgBed
- Multi-storage strategy support: Local, AWS S3, Aliyun OSS, Tencent COS, Qiniu Kodo, Upyun USS, MinIO, WebDAV, SFTP, FTP
- User authentication with JWT and API Key management
- Album and tag system with trash recovery
- Admin dashboard with user groups and permission control
- AI image generation via MiniMax API integration
- Docker Compose deployment support
- RESTful API with usage statistics and rate limiting

## [0.1.0] - 2024-05-25

### Added

- Project scaffolding with Go + Vue 3 architecture
- Basic image upload and retrieval functionality
- SQLite and MySQL database support via GORM
- Frontend built with Vue 3, TypeScript, Tailwind CSS, and shadcn-vue

[Unreleased]: https://github.com/zuquanzhi/sino-imgbed/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/zuquanzhi/sino-imgbed/releases/tag/v0.1.0
