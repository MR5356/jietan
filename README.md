# 介潭
介潭生先龙，先龙生玄鼋，玄鼋生灵龟，灵龟生庶龟，凡介者生于庶龟。 ——《淮南子》

## 架构设计

```mermaid
---
title: Executor
---
classDiagram
    Executor <|-- AsyncExecutor
    Executor <|-- SyncExecutor
    Executor: +Execute()
    class AsyncExecutor {
    }
```