# AGENTS.md

This file provides context for AI coding assistants (Claude Code, Cursor, GitHub Copilot, Codex, etc.) working with this repository.

## 项目简介

该项目是一个身份和访问管理（IAM，Identity and Access Management）系统，旨在为 SaaS 系统提供认证服务。它包括用户注册、登录、权限管理等功能，并支持多租户环境以确保数据隔离和安全性。同时，该系统设计为高可用和可扩展，以满足不断增长的用户需求。

## Development Setup

### Requirements

采用前后端分离的项目架构，技术栈如下：

- **后端**: Golang + Gin 框架
- **前端**: Vue3 + TypeScript
- **数据库**: PostgreSQL
- **缓存**: Redis
- **测试框架**: Python (uv 包管理) + pytest ，端到端测试 (End-to-End Test) 或 黑盒测试 (Black-box Test)
- **UI 自动化测试**: 使用 `agent-browser` 技能（skill）进行页面验证、交互测试等
- **部署**: Docker Compose (需要的中间件，通过 Docker Compose 进行部署)

## 文档索引

## 项目管理

- 在 `docs/PLANS.md` 文件，记录项目的计划和里程碑
- **产品需求文档** (PRD) 在 `docs/prd/README.md` 维护，**设计设计文档** (TDD) 在 `docs/tdd/README.md` 维护。
- 当文档规模过于庞大的时候，可以拆分成多个子文档，放在对应的目录下进行维护，此时目录中的 `README.md` 文件作为索引，链接到各个子文档。

## 开发流程

1. 先用 `using-superpowers` 确认应该先找 skill 再动手。
2. 有新需求或行为变更时，先用 `brainstorming` 做设计澄清，需要产出**产品需求文档（Product Requirements Document, PRD）**详细阐述了产品的功能性需求和非功能性需求，确定了需求后，需要根据prd产出**技术设计文档（Technical Design Document, TDD）**决定如何实现对应的需求。
3. 前一步骤的设计（含prd和tdd两个文档）批准后，用 `writing-plans` 写实施计划。
4. 实施前用 `using-git-worktrees` 建立隔离工作区。
5. 如果平台支持 subagent，优先用 `subagent-driven-development` 按任务执行。
6. 如果不在当前会话里执行，或不能方便使用 subagent，则用 `executing-plans`。
7. 如果涉及前端页面的验证，请使用 `agent-browser` 技能进行
8. 实施过程中配合 `test-driven-development`、`systematic-debugging`、`requesting-code-review`、`receiving-code-review`、`verification-before-completion` 保证质量。
9. 完成后用 `finishing-a-development-branch` 做收尾、合并、PR 或清理。

## Coding Standards

- 使用 docker-compose 启动数据库和相关的中间件
- 使用 Python 对于接口进行端到端测试 (End-to-End Test) 或 黑盒测试 (Black-box Test)，使用 uv 作为 Python 项目管理工具
- Go 的核心代码需要编写适当的单元测试
- 业务代码要多使用依赖助于，提高可维护性，也有助于单元测试的 mock
