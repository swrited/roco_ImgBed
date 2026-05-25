# Contributing to Sino-ImgBed

Thank you for your interest in contributing to Sino-ImgBed. This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bug fix
4. Make your changes
5. Run tests and ensure code quality
6. Submit a pull request

## Development Setup

### Backend

```bash
cd backend-go
go mod download
go run cmd/server/main.go
```

The backend server will start at `http://localhost:8000`.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The development server will start at `http://localhost:5173`.

## Code Style

### Go

- Format all Go code with `gofmt`
- Follow standard Go project structure conventions
- Write meaningful variable and function names
- Add comments for exported functions and types
- Include unit tests for new functionality

### TypeScript / Vue

- Follow the existing ESLint and Prettier configuration
- Use `<script setup>` syntax for Vue components
- Provide complete TypeScript type definitions
- Use composition API patterns consistently

## Commit Message Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

Examples:

```
feat(storage): add WebDAV strategy support
fix(api): resolve upload size limit validation
 docs(readme): update deployment instructions
```

## Pull Request Process

1. Update the README.md or documentation if your changes affect user-facing behavior
2. Ensure all tests pass before submitting
3. Fill out the pull request template completely
4. Link any related issues using keywords (e.g., "Fixes #123")
5. Wait for review from maintainers

## Reporting Issues

When reporting bugs, please include:

- A clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, browser, Go/Node versions)
- Relevant logs or error messages

## Security Issues

Please do not report security vulnerabilities in public issues. See [SECURITY.md](./SECURITY.md) for responsible disclosure guidelines.

## License

By contributing to Sino-ImgBed, you agree that your contributions will be licensed under the MIT License.
