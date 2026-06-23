package errcode

// 定义规则 以1000分段 1000前为系统错误
const (
	Ok         uint32 = 0
	System     uint32 = 100 //系统错误
	EncodeFail uint32 = 101 //封包失败

)

// Login 登录
const (
	LoginSdkFail        uint32 = 1001 //SDK认证失败
	LoginAccountErr     uint32 = 1002 //账号不存在
	LoginSelectRoles    uint32 = 1003 //无法查询账号的角色列表
	LoginCreateRoleMax  uint32 = 1004 //创建角色已经超上限
	LoginCreateRoleFail uint32 = 1005 //创建角色失败
	LoginRoleNameExists uint32 = 1006 //角色名已存在
	LoginRoleNameEmpty  uint32 = 1006 //角色名为空

	LoginGenderNotExists uint32 = 1007 //不存在该性别
	LoginCareerNotExists uint32 = 1008 //不存在该职业
	LoginInvalidRoleID   uint32 = 1009 //角色id不对
	LoginRoleNotExists   uint32 = 1010 //角色不存在
	LoginRoleWorkerFail  uint32 = 1011 //角色协程创建失败
	LoginRepeat          uint32 = 1012 //角色重复登录
)

const (
	SceneRoleIDIllegal  uint32 = 2001 //角色ID非法
	SceneLocationFail   uint32 = 2002 //获取角色场景定位失败
	SceneAlreadyIn      uint32 = 2003 //已经在场景里了
	SceneLeaveFail      uint32 = 2004 //退出场景失败
	SceneNonExistent    uint32 = 2004 //场景不存在
	SceneMapNonExistent uint32 = 2005 //大地图不存在
)
