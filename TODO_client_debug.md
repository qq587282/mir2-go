# 客户端调试进度 (2026-04-15)

## 已修复的问题

### 1. RunGate 6Bit 编码优先级错误
- **问题**: `(buffer>>bits)&0x3F+0x3C` 会先加后与
- **修复**: 改为 `(buffer>>bits)&0x3F)+0x3C`
- **位置**: cmd/rungate/main.go:211, 216

### 2. RunGate forwardM2ToClient 消息处理问题
- **问题**: 只处理 20+ 字节的消息，短消息被忽略
- **修复**: 重写消息解析逻辑，正确处理 RUNGATECODE 格式
- **位置**: cmd/rungate/main.go:166-199

### 3. M2Server sendCharInfo 格式错误
- **问题**: 直接发送编码后的字符串，没有使用 RUNGATECODE 格式
- **修复**: 使用 `sess.Send(network.EncodePacket(msg))`
- **位置**: cmd/m2server/main.go:414-425

### 4. M2Server 对 RunGate 连接发送 SM_SERVERCONFIG
- **问题**: onConnect 时发送 SM_SERVERCONFIG，但此时客户端未注册
- **修复**: 移除 onConnect 中的 sendServerVersion，在 handleGM_OPEN 后发送
- **位置**: cmd/m2server/main.go:150-163, 240-245

### 5. RunGate processBuffer 缺少 continue
- **问题**: 处理完 #...! 消息后没有跳过继续处理
- **修复**: 添加 continue 语句
- **位置**: pkg/network/gate.go:199

### 6. handleLoginServerConnection 使用 nil cfg
- **问题**: 函数使用全局 cfg 变量，可能为 nil
- **修复**: 传递 cfg 参数给函数
- **位置**: cmd/m2server/main.go:112-115, 1205-1210

### 7. RunGate 客户端消息最小长度检查
- **问题**: 最小长度检查为 14 字节，但 TDefaultMessage 只需 8 字节
- **修复**: 将最小长度改为 8 字节
- **位置**: cmd/rungate/main.go:267

### 8. RunGate 消息 Ident 解析错误 (2026-04-15)
- **问题**: onRunMessage 从 decoded[0:2] 读取 ident，应该是 decoded[4:6]
- **原因**: TDefaultMessage 格式是 Recog(4) + Ident(2) + Param(2) + Tag(2) + Series(2) + Padding(2)
- **修复**: 使用 `binary.LittleEndian.Uint16(decoded[4:6])` 正确读取 ident
- **位置**: cmd/rungate/main.go:293

### 9. mir2_test 客户端 CM_WALK 坐标发送错误 (2026-04-15)
- **问题**: 客户端将 x, y 放入 body，但服务器从 Param/Tag 字段提取坐标
- **修复**: 将 x 放入 msg[6:8] (Param)，y 放入 msg[8:10] (Tag)
- **位置**: cmd/mir2_test/main.go:343-365

### 10. M2Server handleWalk/handleRun 坐标解析错误 (2026-04-15)
- **问题**: 服务器从 param 中提取 x = param & 0xFF, y = (param >> 8) & 0xFF
- **修复**: 直接使用 x = int(param), y = int(tag)
- **位置**: cmd/m2server/main.go:641-672, 674-693

### 11. M2Server broadcastMessage 未编码 (2026-04-15)
- **问题**: broadcastMessage 直接发送 TDefaultMessage，未加上 RUNGATECODE 头
- **修复**: 使用 `network.EncodePacket(data)` 编码后再广播
- **位置**: cmd/m2server/main.go:994-1005

### 12. M2Server onMessage 不处理直接连接 (2026-04-15)
- **问题**: onMessage 要求数据至少 20 字节，但直接连接只发送 14 字节
- **修复**: 
  - 改为检查 len(data) >= 14
  - 先检查是否有 RUNGATECODE 头
  - 无头时直接解析 TDefaultMessage
- **位置**: cmd/m2server/main.go:196-247

### 13. 默认地图未创建 (2026-04-15)
- **问题**: 地图文件不存在时，LoadFromFile 返回错误导致地图未加载
- **修复**: 创建 createDefaultMap 函数，生成 1000x1000 的可行走地图
- **位置**: pkg/game/map/gamemap.go:109-133

## 测试结果 (2026-04-15)

### 直接连接测试结果
```
✅ CM_QUERYCHR (100) - SM_QUERYCHR (520) 返回
✅ CM_SELCHR (103) - SM_STARTPLAY (525), SM_LOGON (50), SM_NEWMAP (51) 返回
✅ CM_TURN (3010) - SM_TURN (10) 返回
✅ CM_WALK (3011) - SM_WALK (11) 返回，坐标 x=289, y=618
✅ CM_RUN (3013) - SM_RUN (13) 返回，坐标 x=289, y=619
✅ CM_HIT (3014) - SM_HIT (14) 返回
```

## 当前状态

### 客户端消息解码问题
客户端发送的消息解码后得到：
- 解码长度: 12 字节
- 解码内容: `000000006b00000000000000`
- 消息标识: 0x6B (107) = SM_CHARLIST (服务器消息)

**问题分析**:
1. 客户端消息格式: `#13<<<Bh<<<<<<<<<<<!`
2. 6Bit 解码: 值 [1, 3, 0, 0, 0, 42, 44, 0, 0, 0, 0, 0]
3. 消息标识 (bytes 4-6): 0x6B (107) = SM_CHARLIST

这表明：
1. 客户端发送的消息不是 CM_QUERYCHR (100)
2. 消息可能是损坏的或使用了不同的协议
3. 客户端版本可能与服务器不兼容

### CM_WALK 无响应问题分析 (已修复)
**根本原因**: 
1. RunGate onRunMessage 从 decoded[0:2] 读取 ident，但 TDefaultMessage 的 ident 在 decoded[4:6]
2. mir2_test 将 x, y 放入 body，但 M2Server 从 Param/Tag 字段提取坐标
3. M2Server handleWalk 从 param 中错误解析坐标

**修复方案**:
1. RunGate 使用 `binary.LittleEndian.Uint16(decoded[4:6])` 读取 ident
2. mir2_test 将 x 放入 msg[6:8] (Param)，y 放入 msg[8:10] (Tag)
3. M2Server 直接使用 x = int(param), y = int(tag)

### 服务器响应
```
SM_SERVERCONFIG (20002) - 服务器配置
SM_CHARLIST (107) - 角色列表
```

## 测试结果

```
=== Simple RunGate Test ===
Connected to RunGate (7000)
Waiting for SM_SERVERCONFIG...
Received 27 bytes during handshake
Sending CM_SELCHR: #1<<<<<BX<<<<<<<<<C\=PUSIpPBm]ZRQn<<<<<<<<<<<<<<<<<<<<<<<<<<<
Read 0: 22 bytes
  SM_STARTPLAY - Character selection success
Read 1: 22 bytes
  SM_LOGON - Login success
Read 2: 22 bytes
  SM_NEWMAP - New map loaded
```

## 消息 ID 参考
- CM_QUERYCHR = 100 (客户端查询角色列表)
- CM_SELCHR = 103 (客户端选择角色)
- SM_CHARLIST = 107 (服务器发送角色列表)
- SM_STARTPLAY = 525
- SM_LOGON = 50
- SM_NEWMAP = 51
- SM_SERVERCONFIG = 20002
- CLIENT_VERSION_NUMBER = 134729835

## 消息 ID 修正

### SM_QUERYCHR vs SM_CHARLIST
- **当前使用**: SM_QUERYCHR = 520
- **部分客户端期望**: SM_CHARLIST = 107
- **说明**: 不同版本的 Mir2 使用不同的消息 ID

### 协议流程
1. 客户端连接 LoginGate (7000)
2. LoginGate 转发到 SelGate (7100)
3. 客户端选择服务器后连接 RunGate (7000)  
4. RunGate 转发到 M2Server (16000)
5. M2Server 处理角色相关消息

### 当前问题
- test_client 发送 CM_IDPASSWORD (2001) 到 RunGate，但应该发送到 LoginGate
- SM_CHARLIST (107) 用于某些客户端版本

## 客户端流程测试结果 (2026-04-15)

### 测试流程 (全部通过)
```
1. LoginSrv (15500) - ✅ 握手成功
2. SM_SERVERNAME - ✅ 服务器列表
3. SM_SELECTSERVER_OK - ✅ 服务器选择
4. SM_LOGINRESULT - ✅ 登录验证

5. RunGate (7200) - ✅ 连接成功
6. SM_SERVERCONFIG (20002) - ✅ 服务器配置

7. CM_QUERYCHR (100) - ✅ SM_QUERYCHR (520)
8. CM_SELCHR (103) - ✅ SM_STARTPLAY (525), SM_LOGON (50), SM_NEWMAP (51)
9. CM_TURN (3010) - ✅ SM_TURN (10)
10. CM_WALK (3011) - ✅ SM_WALK (11), 坐标 x=289, y=618
11. CM_RUN (3013) - ✅ SM_RUN (13), 坐标 x=289, y=619
12. CM_HIT (3014) - ✅ SM_HIT (14)
13. CM_SPELL (3015) - 🔄 待实现
14. CM_PICKUP (3018) - 🔄 待实现
```

### 端口配置
- RunGate: 7200 -> M2Server: 16000

## TDefaultMessage 格式 (14字节)

```
偏移  大小  字段     说明
0     4     Recog   标识符
4     2     Ident   消息ID
6     2     Param   参数1 (x坐标)
8     2     Tag     参数2 (y坐标)
10    2     Series  参数3
12    2     -       填充
```

## 待实现功能
- CM_SPELL (3015) - 施法指令
- CM_PICKUP (3018) - 拾取物品
