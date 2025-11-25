export namespace config {
	
	export class Accounts {
	    SoopliveAccount: string;
	    SooplivePassword: string;
	    FlextvAccount: string;
	    FlextvPassword: string;
	    PopkontvAccount: string;
	    PartnerCode: string;
	    PopkontvPassword: string;
	    TwitcastingType: string;
	    TwitcastingAccount: string;
	    TwitcastingPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new Accounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SoopliveAccount = source["SoopliveAccount"];
	        this.SooplivePassword = source["SooplivePassword"];
	        this.FlextvAccount = source["FlextvAccount"];
	        this.FlextvPassword = source["FlextvPassword"];
	        this.PopkontvAccount = source["PopkontvAccount"];
	        this.PartnerCode = source["PartnerCode"];
	        this.PopkontvPassword = source["PopkontvPassword"];
	        this.TwitcastingType = source["TwitcastingType"];
	        this.TwitcastingAccount = source["TwitcastingAccount"];
	        this.TwitcastingPassword = source["TwitcastingPassword"];
	    }
	}
	export class Authorization {
	    PopkontvToken: string;
	
	    static createFrom(source: any = {}) {
	        return new Authorization(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PopkontvToken = source["PopkontvToken"];
	    }
	}
	export class Cookies {
	    Douyin: string;
	    Kuaishou: string;
	    Tiktok: string;
	    Huya: string;
	    Douyu: string;
	    Yy: string;
	    Bilibili: string;
	    Xiaohongshu: string;
	    Bigo: string;
	    Blued: string;
	    Sooplive: string;
	    Netease: string;
	    Qiandu: string;
	    Pandatv: string;
	    Maoerfm: string;
	    Winktv: string;
	    Flextv: string;
	    Look: string;
	    Twitcasting: string;
	    Baidu: string;
	    Weibo: string;
	    Kugou: string;
	    Twitch: string;
	    Liveme: string;
	    Huajiao: string;
	    Liuxing: string;
	    Showroom: string;
	    Acfun: string;
	    Changliao: string;
	    Yinbo: string;
	    Yingke: string;
	    Zhihu: string;
	    Chzzk: string;
	    Haixiu: string;
	    Vvxqiu: string;
	    Live17: string;
	    Langlive: string;
	    Pplive: string;
	    Room6: string;
	    Lehaitv: string;
	    Huamao: string;
	    Shopee: string;
	    Youtube: string;
	    Taobao: string;
	    Jd: string;
	    Faceit: string;
	    Migu: string;
	    Lianjie: string;
	    Laixiu: string;
	    Picarto: string;
	
	    static createFrom(source: any = {}) {
	        return new Cookies(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Douyin = source["Douyin"];
	        this.Kuaishou = source["Kuaishou"];
	        this.Tiktok = source["Tiktok"];
	        this.Huya = source["Huya"];
	        this.Douyu = source["Douyu"];
	        this.Yy = source["Yy"];
	        this.Bilibili = source["Bilibili"];
	        this.Xiaohongshu = source["Xiaohongshu"];
	        this.Bigo = source["Bigo"];
	        this.Blued = source["Blued"];
	        this.Sooplive = source["Sooplive"];
	        this.Netease = source["Netease"];
	        this.Qiandu = source["Qiandu"];
	        this.Pandatv = source["Pandatv"];
	        this.Maoerfm = source["Maoerfm"];
	        this.Winktv = source["Winktv"];
	        this.Flextv = source["Flextv"];
	        this.Look = source["Look"];
	        this.Twitcasting = source["Twitcasting"];
	        this.Baidu = source["Baidu"];
	        this.Weibo = source["Weibo"];
	        this.Kugou = source["Kugou"];
	        this.Twitch = source["Twitch"];
	        this.Liveme = source["Liveme"];
	        this.Huajiao = source["Huajiao"];
	        this.Liuxing = source["Liuxing"];
	        this.Showroom = source["Showroom"];
	        this.Acfun = source["Acfun"];
	        this.Changliao = source["Changliao"];
	        this.Yinbo = source["Yinbo"];
	        this.Yingke = source["Yingke"];
	        this.Zhihu = source["Zhihu"];
	        this.Chzzk = source["Chzzk"];
	        this.Haixiu = source["Haixiu"];
	        this.Vvxqiu = source["Vvxqiu"];
	        this.Live17 = source["Live17"];
	        this.Langlive = source["Langlive"];
	        this.Pplive = source["Pplive"];
	        this.Room6 = source["Room6"];
	        this.Lehaitv = source["Lehaitv"];
	        this.Huamao = source["Huamao"];
	        this.Shopee = source["Shopee"];
	        this.Youtube = source["Youtube"];
	        this.Taobao = source["Taobao"];
	        this.Jd = source["Jd"];
	        this.Faceit = source["Faceit"];
	        this.Migu = source["Migu"];
	        this.Lianjie = source["Lianjie"];
	        this.Laixiu = source["Laixiu"];
	        this.Picarto = source["Picarto"];
	    }
	}
	export class PushSettings {
	    PushChannels: string;
	    DingTalkUrl: string;
	    WeChatUrl: string;
	    BarkUrl: string;
	    BarkLevel: string;
	    BarkSound: string;
	    DingTalkPhone: string;
	    DingTalkAtAll: string;
	    TgToken: string;
	    TgChatId: string;
	    SmtpServer: string;
	    SmtpSsl: string;
	    SmtpPort: string;
	    EmailAccount: string;
	    EmailPassword: string;
	    SenderEmail: string;
	    SenderName: string;
	    ReceiverEmail: string;
	    NtfyUrl: string;
	    NtfyTag: string;
	    NtfyEmail: string;
	    PushPlusToken: string;
	    CustomTitle: string;
	    CustomStartContent: string;
	    CustomStopContent: string;
	    PushOnlyNoRecord: string;
	    PushInterval: number;
	    PushOnStart: string;
	    PushOnStop: string;
	
	    static createFrom(source: any = {}) {
	        return new PushSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PushChannels = source["PushChannels"];
	        this.DingTalkUrl = source["DingTalkUrl"];
	        this.WeChatUrl = source["WeChatUrl"];
	        this.BarkUrl = source["BarkUrl"];
	        this.BarkLevel = source["BarkLevel"];
	        this.BarkSound = source["BarkSound"];
	        this.DingTalkPhone = source["DingTalkPhone"];
	        this.DingTalkAtAll = source["DingTalkAtAll"];
	        this.TgToken = source["TgToken"];
	        this.TgChatId = source["TgChatId"];
	        this.SmtpServer = source["SmtpServer"];
	        this.SmtpSsl = source["SmtpSsl"];
	        this.SmtpPort = source["SmtpPort"];
	        this.EmailAccount = source["EmailAccount"];
	        this.EmailPassword = source["EmailPassword"];
	        this.SenderEmail = source["SenderEmail"];
	        this.SenderName = source["SenderName"];
	        this.ReceiverEmail = source["ReceiverEmail"];
	        this.NtfyUrl = source["NtfyUrl"];
	        this.NtfyTag = source["NtfyTag"];
	        this.NtfyEmail = source["NtfyEmail"];
	        this.PushPlusToken = source["PushPlusToken"];
	        this.CustomTitle = source["CustomTitle"];
	        this.CustomStartContent = source["CustomStartContent"];
	        this.CustomStopContent = source["CustomStopContent"];
	        this.PushOnlyNoRecord = source["PushOnlyNoRecord"];
	        this.PushInterval = source["PushInterval"];
	        this.PushOnStart = source["PushOnStart"];
	        this.PushOnStop = source["PushOnStop"];
	    }
	}
	export class RecordingSettings {
	    Language: string;
	    SkipProxyCheck: string;
	    SavePath: string;
	    SplitByAuthor: string;
	    SplitByTime: string;
	    SplitByTitle: string;
	    IncludeTitleInName: string;
	    RemoveEmoji: string;
	    VideoFormat: string;
	    VideoQuality: string;
	    UseProxy: string;
	    ProxyAddress: string;
	    MaxThreads: number;
	    LoopInterval: number;
	    UrlQueueDelay: number;
	    ShowLoopSeconds: string;
	    ShowStreamUrl: string;
	    SplitRecording: string;
	    ForceHttps: string;
	    DiskSpaceThreshold: number;
	    SplitDuration: number;
	    AutoConvertToMp4: string;
	    ReencodeH264: string;
	    DeleteOriginalAfterConvert: string;
	    GenerateTimeSubtitle: string;
	    RunCustomScript: string;
	    CustomScriptCmd: string;
	    ProxyPlatforms: string;
	    ExtraProxyPlatforms: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordingSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Language = source["Language"];
	        this.SkipProxyCheck = source["SkipProxyCheck"];
	        this.SavePath = source["SavePath"];
	        this.SplitByAuthor = source["SplitByAuthor"];
	        this.SplitByTime = source["SplitByTime"];
	        this.SplitByTitle = source["SplitByTitle"];
	        this.IncludeTitleInName = source["IncludeTitleInName"];
	        this.RemoveEmoji = source["RemoveEmoji"];
	        this.VideoFormat = source["VideoFormat"];
	        this.VideoQuality = source["VideoQuality"];
	        this.UseProxy = source["UseProxy"];
	        this.ProxyAddress = source["ProxyAddress"];
	        this.MaxThreads = source["MaxThreads"];
	        this.LoopInterval = source["LoopInterval"];
	        this.UrlQueueDelay = source["UrlQueueDelay"];
	        this.ShowLoopSeconds = source["ShowLoopSeconds"];
	        this.ShowStreamUrl = source["ShowStreamUrl"];
	        this.SplitRecording = source["SplitRecording"];
	        this.ForceHttps = source["ForceHttps"];
	        this.DiskSpaceThreshold = source["DiskSpaceThreshold"];
	        this.SplitDuration = source["SplitDuration"];
	        this.AutoConvertToMp4 = source["AutoConvertToMp4"];
	        this.ReencodeH264 = source["ReencodeH264"];
	        this.DeleteOriginalAfterConvert = source["DeleteOriginalAfterConvert"];
	        this.GenerateTimeSubtitle = source["GenerateTimeSubtitle"];
	        this.RunCustomScript = source["RunCustomScript"];
	        this.CustomScriptCmd = source["CustomScriptCmd"];
	        this.ProxyPlatforms = source["ProxyPlatforms"];
	        this.ExtraProxyPlatforms = source["ExtraProxyPlatforms"];
	    }
	}
	export class Configuration {
	    RecordingSettings: RecordingSettings;
	    PushSettings: PushSettings;
	    Cookies: Cookies;
	    Authorization: Authorization;
	    Accounts: Accounts;
	
	    static createFrom(source: any = {}) {
	        return new Configuration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RecordingSettings = this.convertValues(source["RecordingSettings"], RecordingSettings);
	        this.PushSettings = this.convertValues(source["PushSettings"], PushSettings);
	        this.Cookies = this.convertValues(source["Cookies"], Cookies);
	        this.Authorization = this.convertValues(source["Authorization"], Authorization);
	        this.Accounts = this.convertValues(source["Accounts"], Accounts);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

