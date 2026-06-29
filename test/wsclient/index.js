let ws = null;
let globalLoginData = {};
let currentSelectRole = null;
let heartbeatTimer = null; // 心跳定时器
const HEARTBEAT_INTERVAL = 10000; // 每10秒发送一次心跳
// const WS_URL = "ws://192.168.0.116:8889/ws";
const WS_URL = `ws://${window.location.host}:8889/ws`;

// ===================== 精灵图系统配置 =====================
const SPRITE_CONFIG = {
    // ---------- 必填（根据你的图片修改） ----------
    IMAGE_URL: '战士.png', // 你的精灵图文件名
    CAREER_IMAGES: {
        1: '战士.png',    // 职业1 战士
        2: '法师.png',     // 职业2 法师
        3: '弓箭手.png',     // 职业2 法师
        4: '牧师.png',     // 职业2 法师

        // 更多职业可根据需求添加
    },
    SPRITE_SIZE: 32,          // 每帧宽高（宽/列数）
    COLS: 4,                  // 横向帧数（宽/每帧宽）
    ROWS: 4,                  // 纵向行数（高/每帧高）
    // ---------- 方向映射（根据实际行顺序调整） ----------
    DIR_INDEX: {
        DOWN: 0,   // 第0行朝下
        LEFT: 1,   // 第1行朝左
        RIGHT: 2,  // 第2行朝右
        UP: 3      // 第3行朝上
    },
    // ---------- 其他 ----------
    FRAME_INTERVAL: 150,       // 动画切换间隔(ms)
    SCALE: 1.5,                // 角色放大倍数
    ANCHOR_X: 0.5,             // 水平锚点（0~1）
    ANCHOR_Y: 0.5,             // 垂直锚点（0~1），0.5居中，0.8脚底
    DEBUG: false,              // 是否显示调试边框和标签
};

// 方向常量（方便引用）
const Direction = SPRITE_CONFIG.DIR_INDEX;

// ===================== 角色精灵图类 =====================
class RoleSprite {
    constructor(imageUrl, color = '#f87171') {
        this.imageUrl = imageUrl;
        this.color = color;
        this.image = new Image();
        this.image.src = imageUrl;
        this.imageLoaded = false;
        this.loadError = false;

        this.currentFrame = 0;                 // 当前动画帧
        this.direction = Direction.DOWN;      // 当前方向
        this.isMoving = false;                // 是否移动（控制动画）
        this.frameTimer = 0;
        this.lastUpdate = Date.now();

        this.image.onload = () => {
            this.imageLoaded = true;
            console.log('✅ 精灵图加载成功:', this.image.width, 'x', this.image.height);
        };
        this.image.onerror = () => {
            this.loadError = true;
            console.error('❌ 精灵图加载失败，请检查路径:', imageUrl);
        };
    }
    changeImage(newUrl) {
        if (this.imageUrl === newUrl) return;
        this.imageUrl = newUrl;
        this.imageLoaded = false;
        this.loadError = false;
        this.image = new Image();
        this.image.src = newUrl;
        this.image.onload = () => {
            this.imageLoaded = true;
            console.log('✅ 精灵图切换成功:', newUrl);
        };
        this.image.onerror = () => {
            this.loadError = true;
            console.error('❌ 精灵图切换失败:', newUrl);
        };
    }
    // 设置方向（根据dx, dy）
    setDirection(dx, dy) {
        if (dx === 0 && dy === 0) return;
        const dir = SPRITE_CONFIG.DIR_INDEX;
        let newDir;
        if (dy < 0) newDir = dir.UP;
        else if (dy > 0) newDir = dir.DOWN;
        else if (dx < 0) newDir = dir.LEFT;
        else if (dx > 0) newDir = dir.RIGHT;
        // 安全钳制：确保方向在有效范围内
        if (newDir !== undefined && newDir >= 0 && newDir < SPRITE_CONFIG.ROWS) {
            this.direction = newDir;
        } else {
            console.warn('无效方向索引，保持当前方向');
        }
    }

    // 更新动画帧
    update() {
        if (!this.imageLoaded) return;
        const now = Date.now();
        const delta = now - this.lastUpdate;
        this.lastUpdate = now;
        if (this.isMoving) {
            this.frameTimer += delta;
            if (this.frameTimer >= SPRITE_CONFIG.FRAME_INTERVAL) {
                this.frameTimer = 0;
                this.currentFrame = (this.currentFrame + 1) % SPRITE_CONFIG.COLS;
            }
        } else {
            this.currentFrame = 0;
            this.frameTimer = 0;
        }
        // 二次安全：确保帧索引不越界
        if (this.currentFrame >= SPRITE_CONFIG.COLS) this.currentFrame = 0;
        if (this.direction >= SPRITE_CONFIG.ROWS) this.direction = Direction.DOWN;
    }

    // 绘制精灵
    draw(ctx, x, y, scale = 1) {
        if (!this.imageLoaded || this.loadError) {
            this.drawPlaceholder(ctx, x, y, scale);
            return;
        }

        const size = SPRITE_CONFIG.SPRITE_SIZE * scale;
        let srcX = this.currentFrame * SPRITE_CONFIG.SPRITE_SIZE;
        let srcY = this.direction * SPRITE_CONFIG.SPRITE_SIZE;

        // 安全钳制：防止切片越界
        const maxSrcX = this.image.width - SPRITE_CONFIG.SPRITE_SIZE;
        const maxSrcY = this.image.height - SPRITE_CONFIG.SPRITE_SIZE;
        if (srcX > maxSrcX) srcX = maxSrcX;
        if (srcY > maxSrcY) srcY = maxSrcY;
        if (srcX < 0) srcX = 0;
        if (srcY < 0) srcY = 0;

        const anchorX = SPRITE_CONFIG.ANCHOR_X || 0.5;
        const anchorY = SPRITE_CONFIG.ANCHOR_Y || 0.5;
        const drawX = x - size * anchorX;
        const drawY = y - size * anchorY;

        ctx.drawImage(
            this.image,
            srcX, srcY,
            SPRITE_CONFIG.SPRITE_SIZE, SPRITE_CONFIG.SPRITE_SIZE,
            drawX, drawY,
            size, size
        );

        // 调试信息
        if (SPRITE_CONFIG.DEBUG) {
            ctx.strokeStyle = 'red';
            ctx.lineWidth = 1;
            ctx.strokeRect(drawX, drawY, size, size);
            ctx.fillStyle = 'white';
            ctx.font = '14px Arial';
            ctx.fillText(`D=${this.direction} F=${this.currentFrame}`, x, drawY - 10);
        }
    }

    // 占位图形（当图片加载失败时使用）
    drawPlaceholder(ctx, x, y, scale) {
        const size = 32 * scale;
        ctx.fillStyle = '#f87171';
        ctx.fillRect(x - size/2, y - size/2, size, size);
        ctx.fillStyle = 'white';
        ctx.font = '20px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('?', x, y + 8);
    }
}

// ===================== 创建精灵实例 =====================
// 玩家自己的精灵
const playerSprite = new RoleSprite(SPRITE_CONFIG.IMAGE_URL, '#f87171');
// 其他玩家的精灵缓存
const otherPlayerSprites = new Map();

function getOtherPlayerSprite(roleId) {
    if (otherPlayerSprites.has(roleId)) {
        return otherPlayerSprites.get(roleId);
    }
    const colors = ['#4ade80', '#60a5fa', '#fbbf24', '#a78bfa', '#f472b6', '#34d399', '#fb923c'];
    const color = colors[Math.floor(Math.random() * colors.length)];
    const sprite = new RoleSprite(SPRITE_CONFIG.IMAGE_URL, color);
    otherPlayerSprites.set(roleId, sprite);
    return sprite;
}

// ===================== DOM元素 =====================
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

// ===================== Protobuf 协议定义（完整） =====================
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
                scene_role: {
                    fields: {
                        role_id: { type: "uint64", id: 1 },
                        server_id: { type: "uint32", id: 2 },
                        plat_id: { type: "uint32", id: 3 },
                        role_name: { type: "string", id: 4 },
                        career: { type: "uint32", id: 5 },
                        pos: { type: "pos", id: 6 }
                    }
                },
            }
        },
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
        CreateRoleC2S: {
            fields: {
                name: { type: "string", id: 1 },
                server_id: { type: "uint32", id: 2 },
                plat_id: { type: "uint32", id: 3 },
                gender: { type: "uint32", id: 4 },
                career: { type: "uint32", id: 5 }
            }
        },
        CreateRoleS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                role_id: { type: "uint64", id: 2 },
                server_id: { type: "uint32", id: 3 },
                plat_id: { type: "uint32", id: 4 },
                name: { type: "string", id: 5 },
                gender: { type: "uint32", id: 6 },
                career: { type: "uint32", id: 7 },
                lv: { type: "uint32", id: 8 }
            }
        },
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
        LoginEnterC2S: {
            fields: {}
        },
        LoginEnterS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 }
            }
        },
        EnterSceneS2C: {
            fields: {
                common_notice: { type: "common.notice", id: 1 },
                scene_id: { type: "uint32", id: 2 },
                map_id: { type: "uint32", id: 3 },
                pos: { type: "common.pos", id: 4 }
            }
        },
        MovePosC2S: {
            fields: {
                pos: { type: "common.pos", id: 3 }
            }
        },
        MovePosS2C: {
            fields: {
                role_id: { type: "uint64", id: 1 },
                map_id: { type: "uint32", id: 2 },
                pos: { type: "common.pos", id: 3 }
            }
        },
        RoleViewListS2C: {
            fields: {
                type: { type: "uint32", id: 1 },
                scene_id: { type: "uint32", id: 2 },
                map_id: { type: "uint32", id: 3 },
                scene_roles: { type: "common.scene_role", id: 4, repeated: true }
            }
        },
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

// 地图配置表
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

// ===================== 全局状态 =====================
const keyState = { w: false, a: false, s: false, d: false };
const MOVE_STEP = 10;
let moveLoop = null;
let currentSceneData = null;
const scenePlayerMap = new Map();
let selfRoleId = null;

// ===================== 封包工具 =====================
function encodePacket(protoData, cmd) {
    const bodyLen = protoData.length;
    const buffer = new ArrayBuffer(8 + bodyLen);
    const view = new DataView(buffer);
    view.setUint32(0, cmd, false);
    view.setUint32(4, bodyLen, false);
    new Uint8Array(buffer, 8).set(protoData);
    return buffer;
}

function bufferToHex(buf) {
    return Array.from(new Uint8Array(buf))
        .map(b => b.toString(16).padStart(2, "0"))
        .join(" ");
}

// ===================== WebSocket =====================
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
            if (heartbeatTimer) {
                clearInterval(heartbeatTimer);
                heartbeatTimer = null;
                appendLog("🛑 心跳定时器已停止");
            }
        };

        ws.onmessage = (e) => {
            const allBytes = new Uint8Array(e.data);
            const dataView = new DataView(e.data);
            const cmd = dataView.getUint32(0, false);
            const bodyLen = dataView.getUint32(4, false);
            const bodyBuf = allBytes.slice(8, 8 + bodyLen);

            switch (cmd) {
                case 1000: handleHeartbeatResp(bodyBuf); break;
                case 1001: handleLoginResp(bodyBuf); break;
                case 1002: handleRoleListResp(bodyBuf); break;
                case 1003: handleCreateRoleResp(bodyBuf); break;
                case 1004: handleEnterGameResp(bodyBuf); break;
                case 2000: handleLoginEnterResp(bodyBuf); break;
                case 2001: handleEnterSceneResp(bodyBuf); break;
                case 2002: handleMovePosS2C(bodyBuf); break;
                case 2003: handleRoleViewListS2C(bodyBuf); break;
                case 2004: handleRoleViewDelS2C(bodyBuf); break;
                default: appendLog(`⚠️ 未处理协议号:${cmd}`);
            }
        };
    });
}

// ===================== 协议处理函数 =====================

function handleEnterSceneResp(bodyBuf) {
    const resp = EnterSceneS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: 0, err_msg: "" };
    appendLog(`🔍 2001进入场景推送 | 场景ID:${resp.scene_id} 地图ID:${resp.map_id} 玩家坐标X:${resp.pos.x} Y:${resp.pos.y}`);

    if (notice.err_code !== 0) {
        appendLog(`⚠️ 进入场景失败：${notice.err_msg}`);
        return;
    }

    currentSceneData = {
        sceneId: resp.scene_id,
        mapId: resp.map_id,
        playerX: resp.pos.x,
        playerY: resp.pos.y,
        mapConf: MapConfs[resp.map_id]
    };

    loginPage.classList.add("hidden");
    rolePage.classList.add("hidden");
    document.getElementById("scenePage").classList.remove("hidden");

    document.getElementById("sceneTitle").innerText = `当前场景：${currentSceneData.mapConf.Name}`;
    document.getElementById("mapNameText").innerText = currentSceneData.mapConf.Name;
    document.getElementById("playerPosText").innerText = `(${currentSceneData.playerX}, ${currentSceneData.playerY})`;

    renderSceneCanvas();
}

// ===================== 渲染场景（使用精灵图） =====================
function renderSceneCanvas() {
    if (!currentSceneData || !currentSceneData.mapConf) {
        console.warn('场景数据未初始化，跳过渲染');
        return;
    }
    const canvas = document.getElementById("sceneCanvas");
    const ctx = canvas.getContext("2d");
    const conf = currentSceneData.mapConf;
    const worldX = currentSceneData.playerX;
    const worldY = currentSceneData.playerY;
    const blockSize = conf.BlockSize;
    const totalBlock = conf.Width / blockSize;

    ctx.imageSmoothingEnabled = false;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "#374151";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    ctx.save();
    const scale = 0.7;
    ctx.scale(scale, scale);

    const playerBlockX = Math.floor(worldX / blockSize);
    const playerBlockY = Math.floor(worldY / blockSize);
    const localX = worldX - playerBlockX * blockSize;
    const localY = worldY - playerBlockY * blockSize;

    // 绘制地图网格
    for (let blockX = 0; blockX < totalBlock; blockX++) {
        for (let blockY = 0; blockY < totalBlock; blockY++) {
            const canvasX = canvas.width / 2 - localX + (blockX - playerBlockX) * blockSize;
            const canvasY = canvas.height / 2 - localY + (blockY - playerBlockY) * blockSize;

            ctx.fillStyle = "#374151";
            ctx.fillRect(canvasX, canvasY, blockSize, blockSize);

            ctx.strokeStyle = "#4b5563";
            ctx.lineWidth = 1;
            const cellSize = conf.CellSize;
            for (let x = 0; x <= blockSize; x += cellSize) {
                ctx.beginPath();
                ctx.moveTo(canvasX + x, canvasY);
                ctx.lineTo(canvasX + x, canvasY + blockSize);
                ctx.stroke();
            }
            for (let y = 0; y <= blockSize; y += cellSize) {
                ctx.beginPath();
                ctx.moveTo(canvasX, canvasY + y);
                ctx.lineTo(canvasX + blockSize, canvasY + y);
                ctx.stroke();
            }

            const isInAoiGrid = Math.abs(blockX - playerBlockX) <= 1 && Math.abs(blockY - playerBlockY) <= 1;
            ctx.strokeStyle = isInAoiGrid ? "#60a5fa" : "#6b7280";
            ctx.lineWidth = isInAoiGrid ? 2 : 1;
            ctx.strokeRect(canvasX, canvasY, blockSize, blockSize);
        }
    }

    // 中心十字（调试）
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

    const playerSize = SPRITE_CONFIG.SPRITE_SIZE * SPRITE_CONFIG.SCALE;

    // ---- 绘制其他玩家 ----
    scenePlayerMap.forEach(player => {
        if (player.roleId === selfRoleId) return;

        const tarBlockX = Math.floor(player.x / blockSize);
        const tarBlockY = Math.floor(player.y / blockSize);
        const tarLocalX = player.x - tarBlockX * blockSize;
        const tarLocalY = player.y - tarBlockY * blockSize;
        const offsetX = tarBlockX - playerBlockX;
        const offsetY = tarBlockY - playerBlockY;
        if (Math.abs(offsetX) > 1 || Math.abs(offsetY) > 1) return;

        const tarCanvasX = canvas.width / 2 - localX + offsetX * blockSize + tarLocalX;
        const tarCanvasY = canvas.height / 2 - localY + offsetY * blockSize + tarLocalY;

        // 获取或创建精灵
        let sprite = getOtherPlayerSprite(player.roleId);
        const careerImages = SPRITE_CONFIG.CAREER_IMAGES;
        if (careerImages && careerImages[player.career]) {
            const newImageUrl = careerImages[player.career];
            sprite.changeImage(newImageUrl);
        }
        sprite.update();

        // 绘制名字
        ctx.font = "bold 13px Microsoft Yahei";
        ctx.fillStyle = "#ffffff";
        ctx.textAlign = "center";
        ctx.shadowColor = "rgba(0,0,0,0.8)";
        ctx.shadowBlur = 4;
        ctx.fillText(player.name, tarCanvasX, tarCanvasY - playerSize - 6);
        ctx.shadowBlur = 0;

        // 绘制精灵
        sprite.draw(ctx, tarCanvasX, tarCanvasY, SPRITE_CONFIG.SCALE);
    });

    // ---- 绘制本地玩家 ----
    playerSprite.update();
    const canvasCenterX = canvas.width / 2;
    const canvasCenterY = canvas.height / 2;

    // 名字（金色高亮）
    ctx.font = "bold 14px Microsoft Yahei";
    ctx.fillStyle = "#fbbf24";
    ctx.textAlign = "center";
    ctx.shadowColor = "rgba(251,191,36,0.5)";
    ctx.shadowBlur = 10;
    ctx.fillText(currentSelectRole.name, canvasCenterX, canvasCenterY - playerSize - 6);
    ctx.shadowBlur = 0;

    // 绘制本地玩家精灵
    playerSprite.draw(ctx, canvasCenterX, canvasCenterY, SPRITE_CONFIG.SCALE);

    ctx.restore();
}

// ===================== 心跳 =====================
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
    ws.send(encodePacket(bin, 1000));
}

function handleHeartbeatResp(bodyBuf) {
    const resp = HeartbeatS2C.decode(bodyBuf);
    const now = Date.now();
    const rtt = now - Number(resp.client_time);
    appendLog(`💚 心跳响应成功 | 客户端时间:${resp.client_time} | 服务端时间:${resp.server_time} | 往返延迟:${rtt}ms`);
}

// ===================== 登录/角色列表 =====================
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

function handleEnterGameResp(bodyBuf) {
    const resp = EnterGameS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: -1, err_msg: "无错误信息" };
    appendLog(`🔍 进入游戏结果 角色ID:${resp.role_id} 错误码:${notice.err_code} ${notice.err_msg}`);
    if (notice.err_code === 0) {
        appendLog(`🎉 角色【${currentSelectRole.name}】成功进入游戏！即将请求场景信息协议2000`);
        sendLoginEnterReq();
        selfRoleId = String(resp.role_id);
        const career = currentSelectRole.career;
        const careerImages = SPRITE_CONFIG.CAREER_IMAGES;
        if (careerImages && careerImages[career]) {
            const newImageUrl = careerImages[career];
            playerSprite.changeImage(newImageUrl);
        }
        playerSprite.direction = Direction.DOWN;
        playerSprite.isMoving = false;
    }
}

function sendLoginEnterReq() {
    const msg = LoginEnterC2S.create({});
    const bin = LoginEnterC2S.encode(msg).finish();
    appendLog(`📦 2000 进入场景请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin, 2000));
}

function handleLoginEnterResp(bodyBuf) {
    const resp = LoginEnterS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: 0, err_msg: "无错误信息" };
    appendLog(`🔍 2000场景信息拉取结果：错误码:${notice.err_code} 描述:${notice.err_msg}`);
    if (notice.err_code === 0) {
        appendLog(`✅ 客户端完整登录流程执行完毕，启动10秒定时心跳`);
        sendHeartbeat();
        heartbeatTimer = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL);
    }
}

// ===================== 创建角色 =====================
function sendCreateRoleReq() {
    const name = document.getElementById('roleName').value.trim();
    const gender = Number(document.querySelector('input[name="gender"]:checked').value);
    const career = Number(document.getElementById('career').value);
    if (!name) { alert("请输入角色名称"); return; }

    const reqData = {
        name, server_id: globalLoginData.server_id,
        plat_id: globalLoginData.plat_id,
        gender, career
    };
    const err = CreateRoleC2S.verify(reqData);
    if (err) {
        appendLog("⚠️ 创建角色参数错误：" + err);
        return;
    }
    const msg = CreateRoleC2S.create(reqData);
    const bin = CreateRoleC2S.encode(msg).finish();
    appendLog(`📦 1003创建角色请求二进制:${bufferToHex(bin)}`);
    ws.send(encodePacket(bin, 1003));
    submitCreateBtn.disabled = true;
}

function handleCreateRoleResp(bodyBuf) {
    const resp = CreateRoleS2C.decode(bodyBuf);
    const notice = resp.common_notice || { err_code: -1, err_msg: "无错误信息" };
    appendLog(`🔍 创建角色结果：错误码${notice.err_code}，${notice.err_msg}`);
    submitCreateBtn.disabled = false;
    if (notice.err_code === 0) {
        appendLog(`🎉 角色【${resp.name}】创建成功，自动刷新角色列表`);
        createRoleForm.classList.add('hidden');
        roleListBox.classList.remove('hidden');
        createRoleBtn.classList.remove('hidden');
        sendSelectRoleReq();
    }
}

// ===================== 移动系统 =====================
function playerMoveLoop() {
    if (!currentSceneData) return;
    let dx = 0, dy = 0;
    if (keyState.w) dy -= MOVE_STEP;
    if (keyState.s) dy += MOVE_STEP;
    if (keyState.a) dx -= MOVE_STEP;
    if (keyState.d) dx += MOVE_STEP;
    if (dx === 0 && dy === 0) {
        playerSprite.isMoving = false;
        return;
    }

    playerSprite.setDirection(dx, dy);
    playerSprite.isMoving = true;

    currentSceneData.playerX += dx;
    currentSceneData.playerY += dy;
    const conf = currentSceneData.mapConf;
    currentSceneData.playerX = Math.max(0, Math.min(conf.Width, currentSceneData.playerX));
    currentSceneData.playerY = Math.max(0, Math.min(conf.Height, currentSceneData.playerY));

    document.getElementById("playerPosText").textContent = `(${currentSceneData.playerX}, ${currentSceneData.playerY})`;
    sendMovePosC2S(currentSceneData.playerX, currentSceneData.playerY);
    renderSceneCanvas();
}

function sendMovePosC2S(x, y) {
    const reqData = { pos: { x, y } };
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

// ===================== 视野同步 =====================
function handleMovePosS2C(bodyBuf) {
    const resp = MovePosS2C.decode(bodyBuf);
    const rid = String(resp.role_id);
    if (scenePlayerMap.has(rid)) {
        const player = scenePlayerMap.get(rid);
        const dx = resp.pos.x - player.x;
        const dy = resp.pos.y - player.y;
        player.x = resp.pos.x;
        player.y = resp.pos.y;

        const sprite = getOtherPlayerSprite(rid);
        sprite.setDirection(dx, dy);
        sprite.isMoving = true;
        if (sprite.moveTimeout) clearTimeout(sprite.moveTimeout);
        sprite.moveTimeout = setTimeout(() => { sprite.isMoving = false; }, 300);
        renderSceneCanvas();
    }
}

function handleRoleViewListS2C(bodyBuf) {
    try {
        const reader = new protobuf.Reader(bodyBuf);
        const SceneRoleType = root.lookupType("common.scene_role");
        let type = 0, scene_id = 0, map_id = 0, roleList = [];

        while (reader.pos < reader.len) {
            const tag = reader.uint32();
            const fieldNum = tag >>> 3;
            switch (fieldNum) {
                case 1: type = reader.uint32(); break;
                case 2: scene_id = reader.uint32(); break;
                case 3: map_id = reader.uint32(); break;
                case 4: {
                    const len = reader.uint32();
                    const subBuf = reader.buf.slice(reader.pos, reader.pos + len);
                    reader.pos += len;
                    const role = SceneRoleType.decode(subBuf);
                    roleList.push(role);
                    break;
                }
                default: reader.skipType(tag & 7);
            }
        }

        appendLog(`📋 视野玩家列表推送 | 类型:${type === 1 ? '全量刷新' : '增量新增'} 场景ID:${scene_id} 地图ID:${map_id} 玩家数量:${roleList.length}`);

        if (type === 1) {
            scenePlayerMap.clear();
            otherPlayerSprites.clear();
        }

        roleList.forEach(role => {
            const rid = String(role.role_id);
            scenePlayerMap.set(rid, {
                roleId: rid,
                name: role.role_name,
                x: role.pos.x,
                y: role.pos.y,
                career: role.career,
            });
            const sprite = getOtherPlayerSprite(rid);
            sprite.isMoving = false;
            sprite.direction = Direction.DOWN;
            appendLog(`👤 进入视野：角色【${role.role_name}】 ID:${rid} 坐标(${role.pos.x},${role.pos.y})`);
        });

        if (!currentSceneData) {
            appendLog("⚠️ 场景数据未初始化，跳过渲染");
            return;
        }
        renderSceneCanvas();
    } catch (err) {
        appendLog(`❌ 2003 视野玩家包解析异常：${err.message}，数据包长度：${bodyBuf.byteLength}`);
        console.error(err);
    }
}

function handleRoleViewDelS2C(bodyBuf) {
    const resp = RoleViewDelS2C.decode(bodyBuf);
    const rid = String(resp.role_id);
    if (scenePlayerMap.has(rid)) {
        const player = scenePlayerMap.get(rid);
        scenePlayerMap.delete(rid);
        otherPlayerSprites.delete(rid);
        appendLog(`👋 离开视野：角色【${player.name}】 ID:${rid}`);
        renderSceneCanvas();
    }
}

// ===================== 键盘事件 =====================
document.addEventListener('keydown', (e) => {
    if (document.getElementById('scenePage').classList.contains('hidden')) return;
    const key = e.key.toLowerCase();
    if (['w', 'a', 's', 'd'].includes(key)) {
        e.preventDefault();
        keyState[key] = true;
        if (!moveLoop) moveLoop = setInterval(playerMoveLoop, 50);
    }
});

document.addEventListener('keyup', (e) => {
    const key = e.key.toLowerCase();
    if (['w', 'a', 's', 'd'].includes(key)) {
        keyState[key] = false;
        if (!keyState.w && !keyState.a && !keyState.s && !keyState.d) {
            clearInterval(moveLoop);
            moveLoop = null;
            playerSprite.isMoving = false;
            renderSceneCanvas();
        }
    }
});

// 方向键支持
document.addEventListener('keydown', (e) => {
    if (document.getElementById('scenePage').classList.contains('hidden')) return;
    switch (e.key) {
        case 'ArrowUp': keyState.w = true; e.preventDefault(); break;
        case 'ArrowDown': keyState.s = true; e.preventDefault(); break;
        case 'ArrowLeft': keyState.a = true; e.preventDefault(); break;
        case 'ArrowRight': keyState.d = true; e.preventDefault(); break;
    }
    if ((keyState.w || keyState.a || keyState.s || keyState.d) && !moveLoop) {
        moveLoop = setInterval(playerMoveLoop, 50);
    }
});

document.addEventListener('keyup', (e) => {
    switch (e.key) {
        case 'ArrowUp': keyState.w = false; break;
        case 'ArrowDown': keyState.s = false; break;
        case 'ArrowLeft': keyState.a = false; break;
        case 'ArrowRight': keyState.d = false; break;
    }
    if (!keyState.w && !keyState.a && !keyState.s && !keyState.d) {
        clearInterval(moveLoop);
        moveLoop = null;
        playerSprite.isMoving = false;
        renderSceneCanvas();
    }
});

// ===================== UI 事件绑定 =====================
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

createRoleBtn.addEventListener('click', () => {
    createRoleForm.classList.remove('hidden');
    roleListBox.classList.add('hidden');
    createRoleBtn.classList.add('hidden');
});

cancelCreateBtn.addEventListener('click', () => {
    createRoleForm.classList.add('hidden');
    roleListBox.classList.remove('hidden');
    createRoleBtn.classList.remove('hidden');
    document.getElementById('roleName').value = '';
    document.querySelector('input[name="gender"][value="1"]').checked = true;
    document.getElementById('career').value = '1';
});

submitCreateBtn.addEventListener('click', sendCreateRoleReq);

// ===================== 预加载提示 =====================
console.log('🎮 游戏客户端已启动，等待登录...');