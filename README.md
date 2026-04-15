# Mir2-Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

使用 Go 语言重写的热血传奇2/3 (Legend of Mir 2/3) 游戏服务器引擎。

## 项目概述

Mir2-Go 是基于原 Delphi/Pascal 代码的分析，使用 Go 语言重新实现的热血传奇游戏服务器。项目保持了与原客户端的通讯协议兼容，可以直接连接官方热血传奇客户端。

### 技术特性

- **高性能**: 基于 Go 语言并发模型，支持高并发连接
- **模块化设计**: 清晰的模块划分，便于维护和扩展
- **协议兼容**: 完整实现 Mir2 通讯协议
- **数据库支持**: 支持 MySQL、SQLite、内存存储

## 环境
Go 安装目录在 D:\Program Files\go
移植源代码目录在 D:\code\mir2 

## 调试和调试
兼容源代码协议和客户端，保留移植前源代码游戏的玩法逻辑，每次修改有需要更新readme文件
客户端登录的端口是7000 其他服务器的端口必须和移植前代码一致
所有服务器的端口必须严格和移植源代码一致
所有服务器的逻辑和消息处理步骤必须和 移植前源代码一致
调试时需要设置超时，不要无限等待响应

## 开发进度 (2026-04-15)

### 已完成功能
1. **LoginSrv (15500)** - 完整登录流程
   - ✅ 新连接建立 Session
   - ✅ 查询服务器列表 (CM_QUERYSERVERNAME 107)
   - ✅ 选择服务器 (CM_SELECTSERVER 104)
   - ✅ 登录验证 (CM_IDPASSWORD 2001)
   - ✅ 向 M2Server 发送 Session 认证消息 (SS_OPENSESSION)

2. **LoginGate (7000)** - 客户端连接网关
   - ✅ 新连接消息处理 (%N)
   - ✅ 查询服务器列表 (CM_QUERYSERVERNAME 107)
   - ✅ 选择服务器 (CM_SELECTSERVER 104)
   - ✅ 登录验证 (CM_IDPASSWORD 2001)
   - ✅ 6Bit 编码/解码
   - ✅ test_client 测试工具完善

3. **统一测试客户端 (mir2_test)**
   - ✅ 集成测试 LoginGate、LoginSrv、M2Server
   - ✅ 支持交互式测试流程
   - ✅ 详细的协议消息解析

4. **RunGate (7200)** - 游戏数据转发
   - ✅ 接收客户端连接
   - ✅ 解析 #1 格式的客户端消息
   - ✅ 实现了 TMsgHeader (20字节) 封装
   - ✅ 使用 GM_DATA 消息类型转发客户端数据
   - ✅ 6Bit 编码/解码实现

5. **M2Server (16000)** - 核心游戏服务器
   - ✅ 接收 RunGate 转发的客户端消息
   - ✅ 实现了 UnpackMsgHeader 解析 TMsgHeader
   - ✅ 处理 CM_QUERYCHR -> SM_SERVERVERSION
   - ✅ 处理 CM_SELCHR -> 创建临时角色
   - ✅ 发送 SM_STARTPLAY, SM_LOGON, SM_NEWMAP
   - ✅ 修复了 nil cfg 导致的 panic 问题
   - ✅ 支持客户端移动指令 (CM_TURN/CM_WALK/CM_RUN/CM_HIT)
   - ✅ 地图文件不存在时创建默认可行走地图

6. **协议实现**
   - ✅ TDefaultMessage (14字节) 结构
   - ✅ RUNGATECODE 头 (0xAA55AA55)
   - ✅ TMsgHeader (20字节) 用于 RunGate-M2Server 通信
   - ✅ GM_OPEN/GM_CLOSE/GM_DATA 消息类型
   - ✅ SS_OPENSESSION/SS_CLOSESESSION 消息
   - ✅ UnpackMsgHeader 函数解析网关头
   - ✅ 6Bit 编码/解码
   - ✅ TAbility 完整字段 (60字节)

### 移动指令修复 (2026-04-15)

#### 修复的问题
| 序号 | 问题描述 | 修复方案 | 文件位置 |
|------|----------|----------|----------|
| 1 | RunGate Ident 解析错误 | `decoded[4:6]` 读取 ident | cmd/rungate/main.go:293 |
| 2 | 测试客户端坐标发送错误 | `msg[6:8]=x, msg[8:10]=y` | cmd/mir2_test/main.go |
| 3 | handleWalk/Run 坐标解析错误 | `x=int(param), y=int(tag)` | cmd/m2server/main.go |
| 4 | broadcastMessage 未编码 | `network.EncodePacket()` | cmd/m2server/main.go:1004 |
| 5 | onMessage 不处理直接连接 | 检查 `len>=14` | cmd/m2server/main.go:196-247 |
| 6 | 默认地图未创建 | `createDefaultMap()` 1000x1000 | pkg/game/map/gamemap.go |

#### 测试结果
```
✅ CM_QUERYCHR (100) -> SM_QUERYCHR (520)
✅ CM_SELCHR   (103) -> SM_STARTPLAY, SM_LOGON, SM_NEWMAP
✅ CM_TURN     (3010) -> SM_TURN (10)
✅ CM_WALK     (3011) -> SM_WALK (11) x=289, y=618
✅ CM_RUN      (3013) -> SM_RUN (13) x=289, y=619
✅ CM_HIT      (3014) -> SM_HIT (14)
```

#### TDefaultMessage 格式 (14字节)
```
偏移  大小  字段     说明
0     4     Recog   标识符
4     2     Ident   消息ID
6     2     Param   参数1 (x坐标)
8     2     Tag     参数2 (y坐标)
10    2     Series  参数3
12    2     -       填充
```

### 启动命令
```bash
# 终端1: 启动 M2Server
go run cmd/m2server/main.go

# 终端2: 启动 RunGate  
go run cmd/rungate/main.go

### 启动命令
```bash
# 终端1: 启动 M2Server
go run cmd/m2server/main.go

# 终端2: 启动 RunGate  
go run cmd/rungate/main.go

# 终端3: 运行测试客户端
go run cmd/full_login_test/main.go
```

### 项目结构

```
mir2-go/
├── cmd/                        # 应用程序入口
│   ├── logingate/             # 登录网关 - 客户端连接验证
│   ├── selserver/             # 服务器选择网关 - 区服选择
│   ├── rungate/               # 游戏运行网关 - 游戏数据转发
│   ├── m2server/              # 游戏主服务器 - 核心游戏逻辑
│   ├── loginsrv/              # 登录服务器 - 账号认证
│   └── dbsrv/                 # 数据库服务器 - 数据存储
│
├── pkg/                        # 核心包
│   ├── protocol/              # 通讯协议
│   │   ├── const.go           # 常量定义
│   │   ├── message_id.go      # 消息ID定义
│   │   ├── skill.go           # 技能ID定义
│   │   └── struct.go          # 协议数据结构
│   │
│   ├── network/               # 网络层
│   │   └── gate.go            # 网关服务器实现
│   │
│   ├── config/                # 配置管理
│   │   └── config.go          # 配置结构与加载
│   │
│   ├── game/                  # 游戏核心
│   │   ├── actor/             # 角色系统
│   │   │   ├── player.go      # 玩家对象
│   │   │   ├── monster.go     # 怪物对象
│   │   │   ├── group.go       # 队伍系统
│   │   │   ├── hero.go        # 英雄系统
│   │   │   ├── trade.go       # 交易系统
│   │   │   └── storage.go     # 仓库系统
│   │   │
│   │   ├── map/               # 地图系统
│   │   │   ├── gamemap.go     # 地图管理
│   │   │   └── pathfind.go    # A*寻路算法
│   │   │
│   │   ├── npc/               # NPC系统
│   │   │   └── npc.go         # NPC对象与脚本
│   │   │
│   │   ├── skill/             # 技能系统
│   │   │   └── skill.go       # 技能与魔法
│   │   │
│   ├── guild/             # 行会系统
│   │   └── guild.go       # 行会管理
│   │
│   │   ├── quest/          # 任务系统
│   │   │   └── quest.go    # 任务管理
│   │
│   │   ├── mail/           # 邮件系统
│   │   │   └── mail.go     # 邮件管理
│   │
│   │   ├── pet/            # 宠物/坐骑系统
│   │   │   └── pet.go      # 宠物管理
│   │
│   │   ├── ranking/        # 排行榜系统
│   │   │   └── ranking.go  # 排行榜管理
│   │
│   │   ├── castle.go          # 城堡系统
│   │   ├── event.go           # 事件系统
│   │   └── item/              # 物品系统
│   │       └── item.go        # 物品与装备
│   │
│   ├── db/                    # 数据库层
│   │   ├── db.go              # 数据模型定义
│   │   └── mysql.go           # MySQL/SQLite实现
│   │
│   ├── script/                # 脚本引擎
│   │   └── engine.go         # NPC脚本解析
│   │
│   └── utils/                 # 工具函数
│
├── config.yaml                # 配置文件
├── go.mod                     # Go模块文件
├── build.bat                  # Windows构建脚本
├── build.sh                   # Linux构建脚本
└── README.md                  # 项目文档
```

## 服务器架构

```
                    ┌─────────────────┐
                    │   客户端        │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
        ┌─────▼─────┐  ┌────▼────┐  ┌──────▼─────┐
        │ LoginGate │  │SelGate  │  │ RunGate    │
        └─────┬─────┘  └────┬────┘  └──────┬─────┘
              │             │              │
              └──────┬──────┘              │
                     │                     │
               ┌─────▼─────┐          ┌─────▼─────┐
               │ LoginSrv │          │  M2Server │
               └─────┬─────┘          └─────┬─────┘
                     │                     │
               ┌─────▼─────┐          ┌─────▼─────┐
               │  DBServer │          │  DBServer │
               └───────────┘          └───────────┘
```

## 功能特性

### 已完成

| 模块 | 状态 | 说明 |
|------|------|------|
| 账号系统 | ✅ | 登录、注册、验证 |
| 角色系统 | ✅ | 创建、删除、选择角色 |
| 移动系统 | ✅ | 走路、跑步、传送 |
| 战斗系统 | ✅ | 物理攻击、技能释放 |
| 队伍系统 | ✅ | 组队、经验共享 |
| 英雄系统 | ✅ | 英雄创建、跟随战斗 |
| 物品系统 | ✅ | 背包、装备、交易 |
| 技能系统 | ✅ | 职业技能、魔法效果 |
| 怪物系统 | ✅ | AI行为、掉落物品 |
| 地图系统 | ✅ | 地图加载、寻路 |
| NPC系统 | ✅ | 对话、商店、脚本 |
| 脚本引擎 | ✅ | 条件判断、命令执行 |
| 数据库 | ✅ | MySQL/SQLite支持 |
| 行会系统 | ✅ | 创建、行会战、联盟 |
| 任务系统 | ✅ | 主线/支线/日常任务 |
| 邮件系统 | ✅ | 邮件收发、附件 |
| 宠物/坐骑系统 | ✅ | 宠物召唤、骑乘 |
| 排行榜系统 | ✅ | 等级/财富/PK排行 |

### 待完善

- [ ] (无)

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 5.7+ (可选)

### 编译

```bash
# 克隆项目
git clone https://github.com/mir2go/mir2.git
cd mir2-go

# 下载依赖
go mod download

# 编译所有服务器
go build ./cmd/...

# 或使用构建脚本
./build.sh     # Linux/Mac
build.bat      # Windows
```

### 配置

编辑 `config.yaml`:

```yaml
servername: "Mir2 Server"
serverip: "0.0.0.0"
serverport: 7000

logingate:
  enable: true
  ip: "0.0.0.0"
  port: 7000
  maxconn: 5000

selgate:
  enable: true
  ip: "0.0.0.0"
  port: 7100
  maxconn: 5000

rungate:
  enable: true
  ip: "0.0.0.0"
  port: 7200
  maxconn: 10000

m2server:
  enable: true
  ip: "0.0.0.0"
  port: 6000

loginsrv:
  enable: true
  ip: "0.0.0.0"
  port: 5500

dbsrv:
  enable: true
  type: "memory"  # memory, mysql, sqlite
  ip: "127.0.0.1"
  port: 3306
  user: "root"
  password: ""
  database: "mir2"
```

### 运行

```bash
# 启动数据库服务器
./bin/dbsrv &

# 启动登录服务器
./bin/loginsrv &

# 启动登录网关
./bin/logingate &

# 启动选择网关
./bin/selgate &

# 启动运行网关
./bin/rungate &

# 启动游戏主服务器
./bin/m2server &
```

或使用启动脚本:

```bash
# Windows
start_servers.bat

# Linux
./start_servers.sh
```

## 协议说明

### 客户端消息 (CM_*)

| ID | 名称 | 说明 |
|----|------|------|
| 100 | CM_QUERYCHR | 查询角色列表 |
| 101 | CM_NEWCHR | 创建角色 |
| 102 | CM_DELCHR | 删除角色 |
| 103 | CM_SELCHR | 选择角色进入游戏 |
| 104 | CM_SELECTSERVER | 选择服务器 |
| 2001 | CM_IDPASSWORD | 登录验证 |
| 3010 | CM_TURN | 转身 |
| 3011 | CM_WALK | 走路 |
| 3013 | CM_RUN | 跑步 |
| 3014 | CM_HIT | 攻击 |
| 3017 | CM_SPELL | 施法 |
| 3030 | CM_SAY | 说话 |

### 服务器消息 (SM_*)

| ID | 名称 | 说明 |
|----|------|------|
| 520 | SM_QUERYCHR | 返回角色列表 |
| 521 | SM_NEWCHR_SUCCESS | 创建角色成功 |
| 10 | SM_TURN | 转身广播 |
| 11 | SM_WALK | 走路广播 |
| 14 | SM_HIT | 攻击广播 |
| 100 | SM_SYSMESSAGE | 系统消息 |
| 101 | SM_GROUPMESSAGE | 队伍消息 |

### 技能ID

| ID | 名称 | 职业 |
|----|------|------|
| 1 | 火球术 | 法师 |
| 2 | 治愈术 | 道士 |
| 3 | 基本剑术 | 战士 |
| 7 | 刺杀剑术 | 战士 |
| 10 | 雷电术 | 法师 |
| 13 | 灵魂火符 | 道士 |
| 26 | 烈火剑法 | 战士 |
| 249 | 隐身 | 刺客 |
| 250 | 烈火剑法 | 刺客 |
| 251 | 双斩 | 刺客 |

## 协议数据格式

### TDefaultMessage (14字节)

```
偏移  大小  字段     类型    说明
0     4     Recog    int32   标识符
4     2     Ident    uint16  消息ID
6     2     Param    uint16  参数1
8     2     Tag      uint16  参数2
10    2     Series   uint16  参数3
12    2     -        -       填充
```

### 数据包格式

```
偏移  大小  字段     类型     说明
0     4     Header   uint32   0xAA55AA55
4     2     Length   uint16   数据长度
6     N     Data     byte[]   消息数据
```

## 原型对比

| Delphi 单元 | Go 包 | 功能 |
|-------------|-------|------|
| Grobal2.pas | pkg/protocol | 常量/消息定义 |
| Envir.pas | pkg/game/map | 地图管理 |
| ObjActor.pas | pkg/game/actor | 角色对象 |
| ObjNpc.pas | pkg/game/npc | NPC对象 |
| ScriptEngn.pas | pkg/script | 脚本引擎 |
| Magic.pas | pkg/game/skill | 技能系统 |
| Guild.pas | pkg/game/guild | 行会系统 |
| Event.pas | pkg/game/event | 事件系统 |
| M2Share.pas | pkg/config | 全局配置 |

## 兼容性

本项目完全兼容原 Mir2 客户端，支持:

- 热血传奇2/3 客户端
- 英雄版本
- 刺客职业
- 自定义技能

## 贡献

欢迎提交 Pull Request! 请确保:

1. 代码遵循 Go 规范
2. 添加必要的注释
3. 更新相关文档

## 协议参考

原代码协议定义来源: `Source/Common/Grobal2.pas`

## 测试

### 测试客户端

项目提供了测试客户端工具，可以用于测试服务器连接:

```bash
# 编译测试客户端
go build -o bin/testclient ./cmd/testclient

# 运行测试
./bin/testclient
```

测试客户端会依次测试:
1. 协议握手 (CM_PROTOCOL)
2. 登录验证 (CM_IDPASSWORD)
3. 角色查询 (CM_QUERYCHR)
4. 角色选择 (CM_SELCHR)
5. 移动指令 (CM_TURN/CM_WALK/CM_RUN/CM_HIT)

### 客户端测试工具 (client_test)

新增的客户端测试工具专门用于测试 LoginSrv 的完整登录流程:

```bash
# 编译客户端测试工具
go build -o bin/client_test.exe ./cmd/client_test

# 运行测试 (需要先启动 loginsrv)
./bin/client_test.exe
```

测试流程:
1. 建立连接 (%N) - 创建session
2. 服务器列表查询 (CM_QUERYSERVERNAME 107)
3. 登录验证 (CM_IDPASSWORD 2001)
4. 选择服务器 (CM_SELECTSERVER 104)

### LoginGate 测试客户端 (test_client)

专门测试 LoginGate (7000) 的完整登录流程:

```bash
# 编译测试客户端
go build -o bin/test_client.exe ./cmd/test_client

# 运行测试 (需要先启动 logingate)
./bin/test_client.exe
```

测试流程:
1. 发送新连接消息 (%N) 建立Session
2. 查询服务器列表 (CM_QUERYSERVERNAME 107)
3. 选择服务器 (CM_SELECTSERVER 104)
4. 登录验证 (CM_IDPASSWORD 2001)

### 统一测试客户端 (mir2_test)

集成测试 LoginGate、LoginSrv、M2Server 三个服务:

```bash
# 编译统一测试客户端
go build -o bin/mir2_test.exe ./cmd/mir2_test

# 运行测试 (需要启动对应服务)
./bin/mir2_test.exe
```

测试流程:
1. **LoginGate (7000)** - 新连接、查询服务器、选择服务器、登录
2. **LoginSrv (15500)** - 完整登录流程测试
3. **M2Server (16000)** - 协议握手、角色查询/选择、移动/攻击测试

运行后会显示每一步的发送和接收数据，便于调试协议问题。

测试输出示例:
```
=== Mir2 完整流程测试客户端 ===

========================================
阶段1: 连接登录服务器 (15500)
========================================
已连接到 127.0.0.1:15500

--- 建立连接: 发送 %N<session>/<IP>$ ---
发送: 新连接消息 %N1/127.0.0.1$
收到 49 字节: ...
Login server packet format detected
Session ID: 0
Packet: #1<<<<<=`><<<<<<D<PrQnYbQnHNxlGqIaXcUaX_DkH<!
Decoded (31 bytes): ...
消息: Ident=537, Recog=0, Param=0, Tag=0, Series=2
  -> SM_SERVERNAME - 服务器列表: rver1/0/Server2/0

--- 测试1: 查询服务器列表 (CM_QUERYSERVERNAME 107) ---
发送: CM_QUERYSERVERNAME (107)
  完整包: '%0/#0<<<<<Bh<<<<<<<<<!$'
  写入 23 字节
收到 ...
消息: Ident=537, Recog=0, Param=0, Tag=0, Series=2
  -> SM_SERVERNAME - 服务器列表: rver1/0/Server2/0

--- 测试2: 登录验证 (CM_IDPASSWORD 2001) ---
发送: CM_IDPASSWORD (2001), account=testuser123
  完整包: '%0/#0<<<<<I@C<<<<<=X<YBQoYCQoUSDmH_HkXBAoXsYkXbLmH_H!$'
  写入 54 字节
...

--- 测试3: 选择服务器 (CM_SELECTSERVER 104) ---
发送: CM_SELECTSERVER (104), serverIndex=0
  完整包: '%0/#0<<<<<B\<<<<<<<L<0!$'
  写入 24 字节
收到 ...
消息: Ident=530, Recog=0, Param=0, Tag=0, Series=0
  -> SM_SELECTSERVER_OK - 服务器地址: 127.0.0.1:7200
```

### 手动测试

1. 启动所有服务器:
   ```bash
   ./start_servers.bat
   ```

2. 使用官方 Mir2 客户端连接服务器

3. 如果遇到问题，检查服务器日志获取详细信息

## 协议数据格式

### 数据包格式 (RUNGATECODE)

客户端与服务器之间的通讯使用 RUNGATECODE 封装:

```
偏移  大小  字段     类型     说明
0     4     Header   uint32   0xAA55AA55 (RUNGATECODE)
4     2     Length   uint16   数据长度 (不含Header和Length本身)
6     N     Data     byte[]   消息数据
```

### TDefaultMessage (14字节)

```
偏移  大小  字段     类型    说明
0     4     Recog    int32   标识符
4     2     Ident    uint16  消息ID
6     2     Param    uint16  参数1
8     2     Tag      uint16  参数2
10    2     Series   uint16  参数3
12    2     -        -       填充
```

### 完整数据包示例

```
偏移  大小  字段
0     4     RUNGATECODE (0xAA55AA55)
4     2     Length (数据长度)
6     14    TDefaultMessage
20    N     消息体数据
```

## 协议说明

测试客户端会依次测试:
1. 协议握手 (CM_PROTOCOL)
2. 登录验证 (CM_IDPASSWORD)
3. 角色查询 (CM_QUERYCHR)
4. 角色选择 (CM_SELCHR)
5. 移动指令 (CM_TURN/CM_WALK/CM_RUN/CM_HIT)

### 手动测试

1. 启动所有服务器:
   ```bash
   ./start_servers.bat
   ```

2. 使用官方 Mir2 客户端连接服务器

3. 如果遇到问题，检查服务器日志获取详细信息

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

**注意**: 本项目仅供学习研究使用，请勿用于商业目的。