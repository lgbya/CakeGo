let ws = null;
let globalLoginData = {};
let currentSelectRole = null;
let heartbeatTimer = null; // 心跳定时器
const HEARTBEAT_INTERVAL = 10000; // 每10秒发送一次心跳
// const WS_URL = "ws://192.168.0.116:8889/ws";
const WS_URL = `ws://${window.location.host}:8889/ws`;



// DOM元素
const loginPage = document.getElementById('loginPage');
const rolePage = document.getElementById('rolePage');
const roleListBox = document.getElementById('roleListBox');
const logBox = document.getElementById('logBox');
const loginBtn = document.getElementById('loginBtn');
const createRoleBtn = document.getElementById('createRoleBtn');
const createRoleForm = document.getElementById('createRoleForm');
const submitCreateBtn = document.getElementById('submitCreateBtn');
const cancelCreateBtn = document.getElementById('cancelCreateBtn');

// 日志工具
function appendLog(text) {
    const time = new Date().toLocaleTimeString();
    logBox.innerHTML += `[${time}] ${text}\n`;
    logBox.scrollTop = logBox.scrollHeight;
}

// ===================== Protobuf 协议定义（1000心跳、1001~1004、2000场景） =====================
const root = protobuf.Root.fromJSON({
    nested: {
        common: {
            nested: {
                notice: {
                    fields: {
                        err_code: { type: "int32", id: 1 },
                        err_msg: { type: "string", id: 2 }
                    }
                },
                pos: {
                    fields: {
                        x: { type: "int32", id: 1 },
                        y: { type: "int32", id: 2 }
                    }
                },
                // 修正后的场景角色结构体（和后端完全对齐）
                scene_role: {
                    fields: {
                        role_id: { type: "uint64", id: 1 },
                        server_id: { type: "uint32", id: 2 },
                        plat_id: { type: "uint32", id: 3 },
                        role_name: { type: "string", id: 4 },
                        pos: { type: "pos", id: 5 }
                    }
                },
            }
        },
        // 心跳包 1000
        HeartbeatC2S: {
            fields: {
                client_time: { type: "int64", id: 1 }
            }
        },
        HeartbeatS2C: {
            fields: {
                client_time: { type: "int64", id: 1 },
                server_time: { type: "int64", id: 2 }
            }
        },
        // 登录 1001
        AccountAuthC2S: {
            fields: {
                account: { type: "string", id: 1 },
                server_id: { type: "uint32", id: 2 },
                plat_id: { type: "uint32", id: 3 },
                ticket: { type: "string", id: 4 },
                client_version: { type: "string", id: 5 },
                channel_id: { type: "uint32", id: 6 }
            }
        },
        AccountAuthS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                is_auth: { type: "bool", id: 2 }
            }
        },
        // 拉取角色列表 1002
        SelectRolesC2S: {
            fields: {
                account: { type: "string", id: 1 },
                plat_id: { type: "uint32", id: 2 },
                server_id: { type: "uint32", id: 3 }
            }
        },
        SelectRolesS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                role_list: { type: "RoleInfo", id: 2, repeated: true }
            }
        },
        // 角色信息结构体
        RoleInfo: {
            fields: {
                role_id: { type: "uint64", id: 1 },
                server_id: { type: "uint32", id: 2 },
                plat_id: { type: "uint32", id: 3 },
                name: { type: "string", id: 4 },
                gender: { type: "uint32", id: 5 },
                lv: { type: "uint32", id: 6 },
                career: { type: "uint32", id: 7 }
            }
        },
        // 创建角色 1003
        CreateRoleC2S:{
            fields:{
                name:{type:"string",id:1},
                server_id:{type:"uint32",id:2},
                plat_id:{type:"uint32",id:3},
                gender:{type:"uint32",id:4},
                career:{type:"uint32",id:5}
            }
        },
        CreateRoleS2C:{
            fields:{
                common_notice:{type:"common.notice",id:1},
                role_id:{type:"uint64",id:2},
                server_id:{type:"uint32",id:3},
                plat_id:{type:"uint32",id:4},
                name:{type:"string",id:5},
                gender:{type:"uint32",id:6},
                career:{type:"uint32",id:7},
                lv:{type:"uint32",id:8}
            }
        },
        // 选中角色进入游戏 1004
        EnterGameC2S: {
            fields: {
                role_id: { type: "uint64", id: 1 },
                server_id: { type: "uint32", id: 2 },
                plat_id: { type: "uint32", id: 3 }
            }
        },
        EnterGameS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                role_id: { type: "uint64", id: 2 }
            }
        },
        // 登录进入场景 2000
        LoginEnterC2S: {
            fields: {}
        },
        LoginEnterS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 }
            }
        },

        // 进入场景推送 2001
        EnterSceneS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                scene_id: { type: "uint32", id: 2 },
                map_id: { type: "uint32", id: 3 },
                pos: { type: "common.pos", id: 4 }
            }
        },
        // 玩家移动 2002
        MovePosC2S: {
            fields: {
                pos: { type: "common.pos", id: 3 }
            }
        },
        // 2002 玩家移动广播
        MovePosS2C: {
            fields: {
                role_id: { type: "uint64", id: 1 },
                map_id: { type: "uint32", id: 2 },
                pos: { type: "common.pos", id: 3 }
            }
        },
        // 2003 视野玩家列表
        RoleViewListS2C: {
            fields: {
                type: { type: "uint32", id: 1 },
                scene_id: { type: "uint32", id: 2 },
                map_id: { type: "uint32", id: 3 },
                scene_roles: { type: "common.scene_role", id: 4, repeated: true }
            }
        },
        // 2004 视野移除玩家
        RoleViewDelS2C: {
            fields: {
                role_id: { type: "uint64", id: 1 }
            }
        }
    }
});

// 协议构造器
const HeartbeatC2S = root.lookupType("HeartbeatC2S");
const HeartbeatS2C = root.lookupType("HeartbeatS2C");
const AccountAuthC2S = root.lookupType("AccountAuthC2S");
const AccountAuthS2C = root.lookupType("AccountAuthS2C");
const SelectRolesC2S = root.lookupType("SelectRolesC2S");
const SelectRolesS2C = root.lookupType("SelectRolesS2C");
const CreateRoleC2S = root.lookupType("CreateRoleC2S");
const CreateRoleS2C = root.lookupType("CreateRoleS2C");
const EnterGameC2S = root.lookupType("EnterGameC2S");
const EnterGameS2C = root.lookupType("EnterGameS2C");
const LoginEnterC2S = root.lookupType("LoginEnterC2S");
const LoginEnterS2C = root.lookupType("LoginEnterS2C");
const EnterSceneS2C = root.lookupType("EnterSceneS2C");
const MovePosC2S = root.lookupType("MovePosC2S");
const MovePosS2C = root.lookupType("MovePosS2C");
const RoleViewListS2C = root.lookupType("RoleViewListS2C");
const RoleViewDelS2C = root.lookupType("RoleViewDelS2C");
// 地图配置表（和服务端配置对齐）
const MapConfs = {
    1000: {
        ID: 1000,
        Name: "奥利利特尔城",
        Type: 1,
        Width: 1600,
        Height: 1600,
        CellSize: 100,
        BlockSize: 200,
    },
    1001: {
        ID: 1001,
        Name: "风花村",
        Type: 1,
        Width: 1600,
        Height: 1600,
        CellSize: 100,
        BlockSize: 200,
    }
};
const keyState = {
    w: false,
    a: false,
    s: false,
    d: false
};
const MOVE_STEP = 10; // 每次移动步长10像素
let moveLoop = null; // 移动循环定时器
// 全局缓存场景、玩家位置数据
let currentSceneData = null;

// 场景所有在线玩家：key = 角色ID(字符串防止大数精度丢失)
const scenePlayerMap = new Map();
// 保存当前玩家自身role_id，用来区分自己和其他玩家
let selfRoleId = null;

// 封包工具：8字节大端包头（4CMD + 4长度）
function encodePacket(protoData, cmd) {
    const bodyLen = protoData.length;
    const buffer = new ArrayBuffer(8 + bodyLen);
    const view = new DataView(buffer);
    view.setUint32(0, cmd, false);
    view.setUint32(4, bodyLen, false);
    new Uint8Array(buffer, 8).set(protoData);
    return buffer;
}

// 二进制转十六进制
function bufferToHex(buf) {
    return Array.from(new Uint8Array(buf))
        .map(b => b.toString(16).padStart(2, "0"))
        .join(" ");
}

// 初始化WebSocket
function initWebSocket() {
    return new Promise((resolve, reject) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            resolve();
            return;
        }
        ws = new WebSocket(WS_URL);
        ws.binaryType = "arraybuffer";

        ws.onopen = () => {
            appendLog("✅ WebSocket 连接成功");
            resolve();
        };
        ws.onerror = (err) => {
            appendLog("❌ 连接异常：" + err.message);
            reject(err);
        };
        ws.onclose = () => {
            appendLog("🔌 服务端连接已关闭");
            loginBtn.disabled = false;
            // 连接关闭时清理心跳定时器
            if (heartbeatTimer) {
                clearInterval(heartbeatTimer);
                heartbeatTimer = null;
                appendLog("🛑 心跳定时器已停止");
            }
        };

        ws.onmessage = (e) => {
            const allBytes = new Uint8Array(e.data);
            // appendLog(`📩 收到数据包：${bufferToHex(allBytes)}`);

            const dataView = new DataView(e.data);
            const cmd = dataView.getUint32(0, false);
            const bodyLen = dataView.getUint32(4, false);
            const bodyBuf = allBytes.slice(8, 8 + bodyLen);

            switch (cmd) {
                case 1000:
                    handleHeartbeatResp(bodyBuf);
                    break;
                case 1001:
                    handleLoginResp(bodyBuf);
                    break;
                case 1002:
                    handleRoleListResp(bodyBuf);
                    break;
                case 1003:
                    handleCreateRoleResp(bodyBuf);
                    break;
                case 1004:
                    handleEnterGameResp(bodyBuf);
                    break;
                case 2000:
                    handleLoginEnterResp(bodyBuf);
                    break;
                case 2001:
                    handleEnterSceneResp(bodyBuf);
                    break;
                case 2002:
                    handleMovePosS2C(bodyBuf);
                    break;
                case 2003:
                    handleRoleViewListS2C(bodyBuf);
                    break;
                case 2004:
                    handleRoleViewDelS2C(bodyBuf);
                    break;
                default:
                    appendLog(`⚠️ 未处理协议号:${cmd}`);
            }
        };
    });
}

// 处理进入场景推送 2001
function handleEnterSceneResp(bodyBuf) {
    const resp = EnterSceneS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: 0, err_msg: "" };
    appendLog(`🔍 2001进入场景推送 | 场景ID:${resp.scene_id} 地图ID:${resp.map_id} 玩家坐标X:${resp.pos.x} Y:${resp.pos.y}`);

    if (notice.err_code !== 0) {
        appendLog(`⚠️ 进入场景失败：${notice.err_msg}`);
        return;
    }

    // 缓存场景数据
    currentSceneData = {
        sceneId: resp.scene_id,
        mapId: resp.map_id,
        playerX: resp.pos.x,
        playerY: resp.pos.y,
        mapConf: MapConfs[resp.map_id]
    };

    // 隐藏登录、选角页，跳转场景页
    loginPage.classList.add("hidden");
    rolePage.classList.add("hidden");
    document.getElementById("scenePage").classList.remove("hidden");

    console.log(resp.map_id)
    // 填充场景信息
    document.getElementById("sceneTitle").innerText = `当前场景：${currentSceneData.mapConf.Name}`;
    // document.getElementById("sceneIdText").innerText = currentSceneData.sceneId;
    document.getElementById("mapNameText").innerText = currentSceneData.mapConf.Name;
    document.getElementById("playerPosText").innerText = `(${currentSceneData.playerX}, ${currentSceneData.playerY})`;

    // 绘制场景+玩家
    renderSceneCanvas();
}
function renderSceneCanvas() {
    const canvas = document.getElementById("sceneCanvas");
    const ctx = canvas.getContext("2d");
    const conf = currentSceneData.mapConf;
    const worldX = currentSceneData.playerX;
    const worldY = currentSceneData.playerY;
    const blockSize = conf.BlockSize;
    const mapMaxW = conf.Width;
    const mapMaxH = conf.Height;

    // 整张地图总Block行列数量
    const totalBlock = mapMaxW / blockSize;

    ctx.imageSmoothingEnabled = false;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "#374151";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // 保存画布状态
    ctx.save();
    // 左上角缩放，不会偏移消失
    const scale = 0.7;
    ctx.scale(scale, scale);

    // 玩家所在的后端逻辑Block索引
    const playerBlockX = Math.floor(worldX / blockSize);
    const playerBlockY = Math.floor(worldY / blockSize);
    // 玩家在当前Block内的相对偏移坐标
    const localX = worldX - playerBlockX * blockSize;
    const localY = worldY - playerBlockY * blockSize;

    // 遍历渲染整张地图所有Block（4×4全部格子都展示）
    for (let blockX = 0; blockX < totalBlock; blockX++) {
        for (let blockY = 0; blockY < totalBlock; blockY++) {
            // 当前Block世界起始坐标
            const blockWorldX = blockX * blockSize;
            const blockWorldY = blockY * blockSize;

            // 换算为画布坐标：玩家永远居中
            const canvasX = canvas.width / 2 - localX + (blockX - playerBlockX) * blockSize;
            const canvasY = canvas.height / 2 - localY + (blockY - playerBlockY) * blockSize;

            // 填充区块背景
            ctx.fillStyle = "#374151";
            ctx.fillRect(canvasX, canvasY, blockSize, blockSize);

            // 绘制100px最小网格格子
            ctx.strokeStyle = "#4b5563";
            ctx.lineWidth = 1;
            const cellSize = conf.CellSize;
            // 竖网格线
            for (let x = 0; x <= blockSize; x += cellSize) {
                ctx.beginPath();
                ctx.moveTo(canvasX + x, canvasY);
                ctx.lineTo(canvasX + x, canvasY + blockSize);
                ctx.stroke();
            }
            // 横网格线
            for (let y = 0; y <= blockSize; y += cellSize) {
                ctx.beginPath();
                ctx.moveTo(canvasX, canvasY + y);
                ctx.lineTo(canvasX + blockSize, canvasY + y);
                ctx.stroke();
            }

            // 高亮后端广播的3×3九宫格区块
            const isInAoiGrid = Math.abs(blockX - playerBlockX) <= 1 && Math.abs(blockY - playerBlockY) <= 1;
            if (isInAoiGrid) {
                ctx.strokeStyle = "#60a5fa";
                ctx.lineWidth = 2;
            } else {
                ctx.strokeStyle = "#6b7280";
                ctx.lineWidth = 1;
            }
            ctx.strokeRect(canvasX, canvasY, blockSize, blockSize);
        }
    }

    // 绘制当前Block中心点（黄色十字调试线）
    const centerBlockCanvasX = canvas.width / 2 - localX + blockSize / 2;
    const centerBlockCanvasY = canvas.height / 2 - localY + blockSize / 2;
    ctx.strokeStyle = "#facc15";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(centerBlockCanvasX - 30, centerBlockCanvasY);
    ctx.lineTo(centerBlockCanvasX + 30, centerBlockCanvasY);
    ctx.stroke();
    ctx.beginPath();
    ctx.moveTo(centerBlockCanvasX, centerBlockCanvasY - 30);
    ctx.lineTo(centerBlockCanvasX, centerBlockCanvasY + 30);
    ctx.stroke();

    // 仅渲染后端AOI 3×3九宫格内的其他玩家
    const playerSize = 24;
    scenePlayerMap.forEach(player => {
        if (player.roleId === selfRoleId) return;

        const tarBlockX = Math.floor(player.x / blockSize);
        const tarBlockY = Math.floor(player.y / blockSize);
        const tarLocalX = player.x - tarBlockX * blockSize;
        const tarLocalY = player.y - tarBlockY * blockSize;

        const offsetX = tarBlockX - playerBlockX;
        const offsetY = tarBlockY - playerBlockY;
        // 超出后端广播九宫格，不渲染
        if (Math.abs(offsetX) > 1 || Math.abs(offsetY) > 1) return;

        const tarCanvasX = canvas.width / 2 - localX + offsetX * blockSize + tarLocalX;
        const tarCanvasY = canvas.height / 2 - localY + offsetY * blockSize + tarLocalY;

        ctx.font = "bold 13px Microsoft Yahei";
        ctx.fillStyle = "#ffffff";
        ctx.textAlign = "center";
        ctx.fillText(player.name, tarCanvasX, tarCanvasY - playerSize - 6);

        ctx.beginPath();
        ctx.fillStyle = "#4ade80";
        ctx.arc(tarCanvasX, tarCanvasY - playerSize / 2, playerSize / 2, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = "#1d4ed8";
        ctx.fillRect(tarCanvasX - playerSize / 3, tarCanvasY, playerSize * 0.66, playerSize * 0.7);
    });

    // 渲染本地玩家（画布正中心）
    const canvasCenterX = canvas.width / 2;
    const canvasCenterY = canvas.height / 2;
    ctx.font = "bold 14px Microsoft Yahei";
    ctx.fillStyle = "#ffffff";
    ctx.textAlign = "center";
    ctx.fillText(currentSelectRole.name, canvasCenterX, canvasCenterY - playerSize - 6);

    ctx.beginPath();
    ctx.fillStyle = "#f87171";
    ctx.arc(canvasCenterX, canvasCenterY - playerSize / 2, playerSize / 2, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#3b82f6";
    ctx.fillRect(canvasCenterX - playerSize / 3, canvasCenterY, playerSize * 0.66, playerSize * 0.7);
    ctx.restore();
}

function sendHeartbeat() {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        clearInterval(heartbeatTimer);
        heartbeatTimer = null;
        appendLog("⚠️ 连接已断开，终止心跳发送");
        return;
    }
    const clientTime = Date.now();
    const reqData = { client_time: clientTime };
    const verifyErr = HeartbeatC2S.verify(reqData);
    if (verifyErr) {
        appendLog("⚠️ 心跳参数校验失败：" + verifyErr);
        return;
    }
    const msg = HeartbeatC2S.create(reqData);
    const bin = HeartbeatC2S.encode(msg).finish();
    // appendLog(`💓 发送心跳包[1000]，客户端时间戳：${clientTime}`);
    ws.send(encodePacket(bin, 1000));
}

// 处理心跳返回
function handleHeartbeatResp(bodyBuf) {
    const resp = HeartbeatS2C.decode(bodyBuf);
    const now = Date.now();
    const rtt = now - Number(resp.client_time);
    appendLog(`💚 心跳响应成功 | 客户端时间:${resp.client_time} | 服务端时间:${resp.server_time} | 往返延迟:${rtt}ms`);
}

// 登录返回处理
function handleLoginResp(bodyBuf) {
    const resp = AccountAuthS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: -1, err_msg: "无错误信息" };
    appendLog(`🔍 登录结果：${resp.is_auth ? "认证成功" : "认证失败"}`);
    appendLog(`错误码:${notice.err_code}  说明:${notice.err_msg}`);

    if (notice.err_code === 0 && resp.is_auth) {
        loginPage.classList.add('hidden');
        document.getElementById('scenePage').classList.add('hidden');
        rolePage.classList.remove('hidden');
        appendLog("📝 自动发起【1002】请求角色列表");
        sendSelectRoleReq();
    } else {
        loginBtn.disabled = false;
    }
}

// 请求角色列表
function sendSelectRoleReq() {
    const reqData = {
        account: globalLoginData.account,
        plat_id: globalLoginData.plat_id,
        server_id: globalLoginData.server_id
    };
    const verifyErr = SelectRolesC2S.verify(reqData);
    if (verifyErr) {
        appendLog("⚠️ 角色列表请求参数错误：" + verifyErr);
        return;
    }
    const msg = SelectRolesC2S.create(reqData);
    const bin = SelectRolesC2S.encode(msg).finish();
    appendLog(`📦 1002请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin, 1002));
}

function handleRoleListResp(bodyBuf) {
    try {
        const reader = new protobuf.Reader(bodyBuf);
        const RoleInfo = root.lookupType("RoleInfo");
        let roleList = [];
        let commonNotice = { err_code: 0, err_msg: "" };

        while (reader.pos < reader.len) {
            const tag = reader.uint32();
            const fieldNum = tag >>> 3;
            const wireType = tag & 7;

            if (fieldNum === 1) {
                const len = reader.uint32();
                const sub = reader.buf.slice(reader.pos, reader.pos + len);
                reader.pos += len;
                commonNotice = root.lookupType("common.notice").decode(sub);
            } else if (fieldNum === 2) {
                const len = reader.uint32();
                const sub = reader.buf.slice(reader.pos, reader.pos + len);
                reader.pos += len;
                const role = RoleInfo.decode(sub);
                roleList.push(role);
            } else {
                reader.skipType(wireType);
            }
        }

        appendLog(`🔍 角色列表返回，错误码：${commonNotice.err_code}，共${roleList.length}个角色`);
        roleListBox.innerHTML = '';

        if (roleList.length === 0) {
            roleListBox.innerHTML = '<p>当前账号暂无游戏角色，请创建新角色开始游戏</p>';
        }

        roleList.forEach(role => {
            const genderStr = role.gender === 1 ? "男" : "女";
            const careerStr = role.career === 1 ? "战士" : "法师";
            const dom = document.createElement("div");
            dom.className = "role-item";
            dom.innerHTML = `
            <p>角色ID：${role.role_id}</p>
            <p>角色名：${role.name}</p>
            <p>等级：${role.lv} | 性别：${genderStr} | 职业：${careerStr}</p>
            <p>服务器ID：${role.server_id}</p>
        `;
            dom.onclick = () => {
                appendLog(`✅ 选中角色【${role.name}】，发起进入游戏请求(1004)`);
                currentSelectRole = role;
                sendEnterGameReq(role);
            };
            roleListBox.appendChild(dom);
        });

        createRoleBtn.classList.remove("hidden");
        if (roleList.length >= 5) {
            createRoleBtn.disabled = true;
            createRoleBtn.innerText = "最多创建5个角色";
            createRoleBtn.onclick = () => alert("已达到角色数量上限");
        } else {
            createRoleBtn.disabled = false;
            createRoleBtn.innerText = "创建新角色";
            createRoleBtn.onclick = () => createRoleForm.classList.remove("hidden");
        }
    } catch (err) {
        appendLog(`❌ 解析异常：${err.message}`);
        console.error(err);
        roleListBox.innerHTML = "<p>数据解析失败</p>";
        createRoleBtn.classList.remove("hidden");
    }
}

// 发送进入游戏协议 1004
function sendEnterGameReq(role) {
    const reqData = {
        role_id: role.role_id,
        server_id: role.server_id,
        plat_id: role.plat_id
    };
    const err = EnterGameC2S.verify(reqData);
    if (err) {
        appendLog("⚠️ 进入游戏参数校验失败：" + err);
        return;
    }
    const msg = EnterGameC2S.create(reqData);
    const bin = EnterGameC2S.encode(msg).finish();
    appendLog(`📦 1004进入游戏请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin, 1004));
}

// 进入游戏返回处理
function handleEnterGameResp(bodyBuf) {
    const resp = EnterGameS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: -1, err_msg: "无错误信息" };
    appendLog(`🔍 进入游戏结果 角色ID:${resp.role_id} 错误码:${notice.err_code} ${notice.err_msg}`);
    if (notice.err_code === 0) {
        appendLog(`🎉 角色【${currentSelectRole.name}】成功进入游戏！即将请求场景信息协议2000`);
        sendLoginEnterReq();
        selfRoleId = String(resp.role_id);
    }
}

// 发送登录进入场景 2000
function sendLoginEnterReq() {
    const msg = LoginEnterC2S.create({});
    const bin = LoginEnterC2S.encode(msg).finish();
    appendLog(`📦 2000 进入场景请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin, 2000));
}

// 2000协议返回解析，成功后启动10秒心跳
function handleLoginEnterResp(bodyBuf) {
    const resp = LoginEnterS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: 0, err_msg: "无错误信息" };
    appendLog(`🔍 2000场景信息拉取结果：错误码:${notice.err_code} 描述:${notice.err_msg}`);
    if (notice.err_code === 0) {
        appendLog(`✅ 客户端完整登录流程执行完毕，启动10秒定时心跳`);
        // 立即发送一次心跳，之后每10秒循环
        sendHeartbeat();
        heartbeatTimer = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL);
    }
}

// 发送创建角色请求 1003
function sendCreateRoleReq(){
    const name = document.getElementById('roleName').value.trim();
    const gender = Number(document.querySelector('input[name="gender"]:checked').value);
    const career = Number(document.getElementById('career').value);

    if(!name){
        alert("请输入角色名称");
        return;
    }

    const reqData = {
        name:name,
        server_id:globalLoginData.server_id,
        plat_id:globalLoginData.plat_id,
        gender:gender,
        career:career
    };

    const err = CreateRoleC2S.verify(reqData);
    if(err){
        appendLog("⚠️ 创建角色参数错误："+err);
        return;
    }
    const msg = CreateRoleC2S.create(reqData);
    const bin = CreateRoleC2S.encode(msg).finish();
    appendLog(`📦 1003创建角色请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin,1003));
    submitCreateBtn.disabled = true;
}

// 创建角色返回处理
function handleCreateRoleResp(bodyBuf){
    const resp = CreateRoleS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: -1, err_msg: "无错误信息" };
    appendLog(`🔍 创建角色结果：错误码${notice.err_code}，${notice.err_msg}`);
    submitCreateBtn.disabled = false;
    if(notice.err_code === 0){
        appendLog(`🎉 角色【${resp.name}】创建成功，自动刷新角色列表`);
        createRoleForm.classList.add('hidden');
        // 关闭创建面板，恢复上方角色列表与创建按钮
        roleListBox.classList.remove('hidden');
        createRoleBtn.classList.remove('hidden');
        sendSelectRoleReq();
    }
}

// function handleMovePosS2C(bodyBuf) {
//     const resp = MovePosS2C.decode(bodyBuf);
//     appendLog(`📡 玩家移动广播 | 角色ID:${resp.role_id} 地图ID:${resp.map_id} 坐标:(${resp.pos.x},${resp.pos.y})`);
//     // 后续可扩展：判断role_id不是自己，在画布渲染其他玩家
// }

function handleMovePosS2C(bodyBuf) {
    const resp = MovePosS2C.decode(bodyBuf);
    const rid = String(resp.role_id);
    if (scenePlayerMap.has(rid)) {
        const player = scenePlayerMap.get(rid);
        // 更新坐标
        player.x = resp.pos.x;
        player.y = resp.pos.y;
        // appendLog(`🏃 角色【${player.name}】移动到坐标(${resp.pos.x},${resp.pos.y})`);
        renderSceneCanvas();
    }
}

function handleRoleViewListS2C(bodyBuf) {
    try {
        const reader = new protobuf.Reader(bodyBuf);
        const SceneRoleType = root.lookupType("common.scene_role");

        let type = 0;
        let scene_id = 0;
        let map_id = 0;
        let roleList = [];

        while (reader.pos < reader.len) {
            const tag = reader.uint32();
            const fieldNum = tag >>> 3;
            const wireType = tag & 7;

            switch (fieldNum) {
                case 1:
                    // type uint32
                    type = reader.uint32();
                    break;
                case 2:
                    // scene_id uint32
                    scene_id = reader.uint32();
                    break;
                case 3:
                    // map_id uint32
                    map_id = reader.uint32();
                    break;
                case 4:
                    // repeated common.scene_role 嵌套消息
                    const len = reader.uint32();
                    const subBuf = reader.buf.slice(reader.pos, reader.pos + len);
                    reader.pos += len;
                    const role = SceneRoleType.decode(subBuf);
                    roleList.push(role);
                    break;
                default:
                    reader.skipType(wireType);
                    break;
            }
        }
        appendLog(`📋 视野玩家列表推送 | 类型:${type === 1 ? '全量刷新' : '增量新增'} 场景ID:${scene_id} 地图ID:${map_id} 玩家数量:${roleList.length}`);

        if (type === 1) {
            scenePlayerMap.clear();
        }

        roleList.forEach(role => {
            const rid = String(role.role_id);
            scenePlayerMap.set(rid, {
                roleId: rid,
                name: role.role_name,
                x: role.pos.x,
                y: role.pos.y
            });
            appendLog(`👤 进入视野：角色【${role.role_name}】 ID:${rid} 坐标(${role.pos.x},${role.pos.y})`);
        });

// 关键保护：场景未初始化直接返回，不执行渲染
        if (!currentSceneData) {
            appendLog("⚠️ 场景数据未初始化，跳过渲染");
            return;
        }
        renderSceneCanvas();
    } catch (err) {
        appendLog(`❌ 2003   视野玩家包解析异常：${err.message}，数据包长度：${bodyBuf.byteLength}`);
        console.error("2003手动解析错误", err, bodyBuf);
    }
}
function handleRoleViewDelS2C(bodyBuf) {
    const resp = RoleViewDelS2C.decode(bodyBuf);
    const rid = String(resp.role_id);
    if (scenePlayerMap.has(rid)) {
        const player = scenePlayerMap.get(rid);
        scenePlayerMap.delete(rid);
        appendLog(`👋 离开视野：角色【${player.name}】 ID:${rid}`);
        renderSceneCanvas();
    }
}

// 登录点击事件
loginBtn.addEventListener('click', async () => {
    loginBtn.disabled = true;
    appendLog("================ 开始登录流程 ================");
    try {
        await initWebSocket();
        globalLoginData = {
            account: document.getElementById('account').value.trim(),
            server_id: Number(document.getElementById('serverId').value),
            plat_id: 1,
            ticket: "sdk_verify_ticket_123456",
            client_version: "1.0.0",
            channel_id: 1
        };
        const loginMsg = AccountAuthC2S.create(globalLoginData);
        const loginBin = AccountAuthC2S.encode(loginMsg).finish();
        appendLog(`📦 1001登录包二进制:${bufferToHex(loginBin)}`);
        ws.send(encodePacket(loginBin, 1001));
        appendLog("📤 登录数据包已发送");
    } catch (err) {
        appendLog("❌ 登录异常：" + err.message);
        loginBtn.disabled = false;
    }
});

// 打开创建角色表单
createRoleBtn.addEventListener('click',()=>{
    createRoleForm.classList.remove('hidden');
    // 打开创建面板，隐藏上方角色列表和创建按钮
    roleListBox.classList.add('hidden');
    createRoleBtn.classList.add('hidden');
});
// 取消创建
cancelCreateBtn.addEventListener('click',()=>{
    createRoleForm.classList.add('hidden');
    // 恢复角色列表、创建按钮显示
    roleListBox.classList.remove('hidden');
    createRoleBtn.classList.remove('hidden');
    // 重置表单，清空上次填写内容
    document.getElementById('roleName').value = '';
    document.querySelector('input[name="gender"][value="1"]').checked = true;
    document.getElementById('career').value = '1';
});
// 提交创建角色
submitCreateBtn.addEventListener('click',sendCreateRoleReq);

// 键盘按下
document.addEventListener('keydown', (e) => {
    // 不在场景页直接返回，不处理移动
    if(document.getElementById('scenePage').classList.contains('hidden')){
        return;
    }
    const key = e.key.toLowerCase();
    if (['w','a','s','d'].includes(key)) {
        e.preventDefault();
        keyState[key] = true;
        // 启动移动循环（避免多次重复创建定时器）
        if (!moveLoop) {
            moveLoop = setInterval(playerMoveLoop, 50);
        }
    }
});

// 键盘抬起
document.addEventListener('keyup', (e) => {
    const key = e.key.toLowerCase();
    if (['w','a','s','d'].includes(key)) {
        keyState[key] = false;
        // 所有按键松开则停止移动
        if (!keyState.w && !keyState.a && !keyState.s && !keyState.d) {
            clearInterval(moveLoop);
            moveLoop = null;
        }
    }
});

document.addEventListener('keydown', (e) => {
    // 只有场景页面非隐藏时，键盘移动才生效
    if(document.getElementById('scenePage').classList.contains('hidden')){
        return;
    }
    switch(e.key){
        case 'ArrowUp':
            keyState.w = true;
            e.preventDefault();
            break;
        case 'ArrowDown':
            keyState.s = true;
            e.preventDefault();
            break;
        case 'ArrowLeft':
            keyState.a = true;
            e.preventDefault();
            break;
        case 'ArrowRight':
            keyState.d = true;
            e.preventDefault();
            break;
    }
    // 只要任意移动键按下，启动循环
    if ((keyState.w || keyState.a || keyState.s || keyState.d) && !moveLoop) {
        moveLoop = setInterval(playerMoveLoop, 50);
    }
});

document.addEventListener('keyup', (e) => {
    switch(e.key){
        case 'ArrowUp':
            keyState.w = false;
            break;
        case 'ArrowDown':
            keyState.s = false;
            break;
        case 'ArrowLeft':
            keyState.a = false;
            break;
        case 'ArrowRight':
            keyState.d = false;
            break;
    }
    // 所有按键松开停止移动
    if (!keyState.w && !keyState.a && !keyState.s && !keyState.d) {
        clearInterval(moveLoop);
        moveLoop = null;
    }
});

function playerMoveLoop() {
    if (!currentSceneData) return;
    let dx = 0, dy = 0;
    // 计算位移
    if (keyState.w) dy -= MOVE_STEP;
    if (keyState.s) dy += MOVE_STEP;
    if (keyState.a) dx -= MOVE_STEP;
    if (keyState.d) dx += MOVE_STEP;

    if (dx === 0 && dy === 0) return;

    // 更新当前玩家世界坐标
    currentSceneData.playerX += dx;
    currentSceneData.playerY += dy;

    // 边界限制：防止走出地图
    const conf = currentSceneData.mapConf;
    currentSceneData.playerX = Math.max(0, Math.min(conf.Width, currentSceneData.playerX));
    currentSceneData.playerY = Math.max(0, Math.min(conf.Height, currentSceneData.playerY));

    // 更新页面坐标显示
    document.getElementById("playerPosText").textContent = `(${currentSceneData.playerX}, ${currentSceneData.playerY})`;

    // 发送2002移动协议到服务端
    sendMovePosC2S(currentSceneData.playerX, currentSceneData.playerY);

    // 重新渲染九宫格场景
    renderSceneCanvas();
}

// 发送玩家移动协议 2002
function sendMovePosC2S(x, y) {
    const reqData = {
        pos: { x: x, y: y }
    };
    const err = MovePosC2S.verify(reqData);
    if (err) {
        appendLog(`⚠️ 移动协议参数错误：${err}`);
        return;
    }
    const msg = MovePosC2S.create(reqData);
    const bin = MovePosC2S.encode(msg).finish();
    appendLog(`🎮 发送移动协议[2002]，新坐标：(${x}, ${y})`);
    ws.send(encodePacket(bin, 2002));
}