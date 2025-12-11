package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Configuration struct {
	RecordingSettings RecordingSettings `ini:"录制设置"`
	PushSettings      PushSettings      `ini:"推送配置"`
	Cookies           Cookies           `ini:"Cookie"`
	Authorization     Authorization     `ini:"Authorization"`
	Accounts          Accounts          `ini:"账号密码"`
	WindowSettings    WindowSettings    `ini:"窗口设置"`
	UpdateSettings    UpdateSettings    `ini:"更新设置"`
	AISettings        AISettings        `ini:"AI设置"`
}

type UpdateSettings struct {
	CheckUpdateOnStart string `ini:"启动时检查更新(是/否)"`
	UpdateUrl          string `ini:"更新检查地址"`
}

type WindowSettings struct {
	MiniPosX int `ini:"悬浮球X坐标"`
	MiniPosY int `ini:"悬浮球Y坐标"`
	Width    int `ini:"窗口宽度"`
	Height   int `ini:"窗口高度"`
	X        int `ini:"窗口X坐标"`
	Y        int `ini:"窗口Y坐标"`
}

type RecordingSettings struct {
	Language                   string  `ini:"language(zh_cn/en)"`
	SkipProxyCheck             string  `ini:"是否跳过代理检测(是/否)"`
	SavePath                   string  `ini:"直播保存路径(不填则默认)"`
	SplitByAuthor              string  `ini:"保存文件夹是否以作者区分"`
	SplitByTime                string  `ini:"保存文件夹是否以时间区分"`
	SplitByTitle               string  `ini:"保存文件夹是否以标题区分"`
	IncludeTitleInName         string  `ini:"保存文件名是否包含标题"`
	RemoveEmoji                string  `ini:"是否去除名称中的表情符号"`
	VideoFormat                string  `ini:"视频保存格式ts|mkv|flv|mp4|mp3音频|m4a音频"`
	VideoQuality               string  `ini:"原画|超清|高清|标清|流畅"`
	UseProxy                   string  `ini:"是否使用代理ip(是/否)"`
	ProxyAddress               string  `ini:"代理地址"`
	MaxThreads                 int     `ini:"同一时间访问网络的线程数"`
	LoopInterval               int     `ini:"循环时间(秒)"`
	UrlQueueDelay              int     `ini:"排队读取网址时间(秒)"`
	ShowLoopSeconds            string  `ini:"是否显示循环秒数"`
	ShowStreamUrl              string  `ini:"是否显示直播源地址"`
	SplitRecording             string  `ini:"分段录制是否开启"`
	ForceHttps                 string  `ini:"是否强制启用https录制"`
	DiskSpaceThreshold         float64 `ini:"录制空间剩余阈值(gb)"`
	SplitDuration              int     `ini:"视频分段时间(秒)"`
	AutoConvertToMp4           string  `ini:"录制完成后自动转为mp4格式"`
	ReencodeH264               string  `ini:"mp4格式重新编码为h264"`
	DeleteOriginalAfterConvert string  `ini:"追加格式后删除原文件"`
	GenerateTimeSubtitle       string  `ini:"生成时间字幕文件"`
	RunCustomScript            string  `ini:"是否录制完成后执行自定义脚本"`
	CustomScriptCmd            string  `ini:"自定义脚本执行命令"`
	ProxyPlatforms             string  `ini:"使用代理录制的平台(逗号分隔)"`
	ExtraProxyPlatforms        string  `ini:"额外使用代理录制的平台(逗号分隔)"`
	FFmpegPath                 string  `ini:"ffmpeg路径"`
}

type PushSettings struct {
	PushChannels       string `ini:"直播状态推送渠道"`
	DingTalkUrl        string `ini:"钉钉推送接口链接"`
	WeChatUrl          string `ini:"微信推送接口链接"`
	BarkUrl            string `ini:"bark推送接口链接"`
	BarkLevel          string `ini:"bark推送中断级别"`
	BarkSound          string `ini:"bark推送铃声"`
	DingTalkPhone      string `ini:"钉钉通知@对象(填手机号)"`
	DingTalkAtAll      string `ini:"钉钉通知@全体(是/否)"`
	TgToken            string `ini:"tgapi令牌"`
	TgChatId           string `ini:"tg聊天id(个人或者群组id)"`
	SmtpServer         string `ini:"smtp邮件服务器"`
	SmtpSsl            string `ini:"是否使用SMTP服务SSL加密(是/否)"`
	SmtpPort           string `ini:"SMTP邮件服务器端口"`
	EmailAccount       string `ini:"邮箱登录账号"`
	EmailPassword      string `ini:"发件人密码(授权码)"`
	SenderEmail        string `ini:"发件人邮箱"`
	SenderName         string `ini:"发件人显示昵称"`
	ReceiverEmail      string `ini:"收件人邮箱"`
	NtfyUrl            string `ini:"ntfy推送地址"`
	NtfyTag            string `ini:"ntfy推送标签"`
	NtfyEmail          string `ini:"ntfy推送邮箱"`
	PushPlusToken      string `ini:"pushplus推送token"`
	CustomTitle        string `ini:"自定义推送标题"`
	CustomStartContent string `ini:"自定义开播推送内容"`
	CustomStopContent  string `ini:"自定义关播推送内容"`
	PushOnlyNoRecord   string `ini:"只推送通知不录制(是/否)"`
	PushInterval       int    `ini:"直播推送检测频率(秒)"`
	PushOnStart        string `ini:"开播推送开启(是/否)"`
	PushOnStop         string `ini:"关播推送开启(是/否)"`
}

type Cookies struct {
	Douyin      string `ini:"抖音cookie"`
	Kuaishou    string `ini:"快手cookie"`
	Tiktok      string `ini:"tiktok_cookie"`
	Huya        string `ini:"虎牙cookie"`
	Douyu       string `ini:"斗鱼cookie"`
	Yy          string `ini:"yy_cookie"`
	Bilibili    string `ini:"b站cookie"`
	Xiaohongshu string `ini:"小红书cookie"`
	Bigo        string `ini:"bigo_cookie"`
	Blued       string `ini:"blued_cookie"`
	Sooplive    string `ini:"sooplive_cookie"`
	Netease     string `ini:"netease_cookie"`
	Qiandu      string `ini:"千度热播_cookie"`
	Pandatv     string `ini:"pandatv_cookie"`
	Maoerfm     string `ini:"猫耳fm_cookie"`
	Winktv      string `ini:"winktv_cookie"`
	Flextv      string `ini:"flextv_cookie"`
	Look        string `ini:"look_cookie"`
	Twitcasting string `ini:"twitcasting_cookie"`
	Baidu       string `ini:"baidu_cookie"`
	Weibo       string `ini:"weibo_cookie"`
	Kugou       string `ini:"kugou_cookie"`
	Twitch      string `ini:"twitch_cookie"`
	Liveme      string `ini:"liveme_cookie"`
	Huajiao     string `ini:"huajiao_cookie"`
	Liuxing     string `ini:"liuxing_cookie"`
	Showroom    string `ini:"showroom_cookie"`
	Acfun       string `ini:"acfun_cookie"`
	Changliao   string `ini:"changliao_cookie"`
	Yinbo       string `ini:"yinbo_cookie"`
	Yingke      string `ini:"yingke_cookie"`
	Zhihu       string `ini:"zhihu_cookie"`
	Chzzk       string `ini:"chzzk_cookie"`
	Haixiu      string `ini:"haixiu_cookie"`
	Vvxqiu      string `ini:"vvxqiu_cookie"`
	Live17      string `ini:"17live_cookie"`
	Langlive    string `ini:"langlive_cookie"`
	Pplive      string `ini:"pplive_cookie"`
	Room6       string `ini:"6room_cookie"`
	Lehaitv     string `ini:"lehaitv_cookie"`
	Huamao      string `ini:"huamao_cookie"`
	Shopee      string `ini:"shopee_cookie"`
	Youtube     string `ini:"youtube_cookie"`
	Taobao      string `ini:"taobao_cookie"`
	Jd          string `ini:"jd_cookie"`
	Faceit      string `ini:"faceit_cookie"`
	Migu        string `ini:"migu_cookie"`
	Lianjie     string `ini:"lianjie_cookie"`
	Laixiu      string `ini:"laixiu_cookie"`
	Picarto     string `ini:"picarto_cookie"`
}

type Authorization struct {
	PopkontvToken string `ini:"popkontv_token"`
}

type AISettings struct {
	GeminiApiKey string `ini:"gemini_api_key"`
	GeminiModel  string `ini:"gemini_model"`
}

type Accounts struct {
	SoopliveAccount     string `ini:"sooplive账号"`
	SooplivePassword    string `ini:"sooplive密码"`
	FlextvAccount       string `ini:"flextv账号"`
	FlextvPassword      string `ini:"flextv密码"`
	PopkontvAccount     string `ini:"popkontv账号"`
	PartnerCode         string `ini:"partner_code"`
	PopkontvPassword    string `ini:"popkontv密码"`
	TwitcastingType     string `ini:"twitcasting账号类型"`
	TwitcastingAccount  string `ini:"twitcasting账号"`
	TwitcastingPassword string `ini:"twitcasting密码"`
}

// BackupConfig 将配置文件复制到备份目录并加上时间戳
func BackupConfig(configPath string) error {
	backupDir := filepath.Join(filepath.Dir(configPath), "backup_config")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("config_%s.ini", timestamp))

	input, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, input, 0644)
}
