package platform

// 所有支持的平台配置
var PlatformConfigs = map[PlatformType]PlatformConfig{
	// 优先实现平台
	PlatformTaobao: {
		Code:        PlatformTaobao,
		Name:        "淘宝",
		Icon:        "taobao",
		Description: "阿里巴巴旗下C2C电商平台",
		Features:    []string{"orders", "products", "logistics", "refund"},
		AuthType:    "oauth",
	},
	PlatformTmall: {
		Code:        PlatformTmall,
		Name:        "天猫",
		Icon:        "tmall",
		Description: "阿里巴巴旗下B2C电商平台",
		Features:    []string{"orders", "products", "logistics", "refund"},
		AuthType:    "oauth",
	},
	PlatformDouyin: {
		Code:        PlatformDouyin,
		Name:        "抖音电商",
		Icon:        "douyin",
		Description: "抖音直播电商平台",
		Features:    []string{"orders", "products", "logistics"},
		AuthType:    "oauth",
	},
	PlatformKuaishou: {
		Code:        PlatformKuaishou,
		Name:        "快手电商",
		Icon:        "kuaishou",
		Description: "快手直播电商平台",
		Features:    []string{"orders", "products"},
		AuthType:    "oauth",
	},
	PlatformWechatVideo: {
		Code:        PlatformWechatVideo,
		Name:        "微信视频号",
		Icon:        "wechat_video",
		Description: "微信视频号小店",
		Features:    []string{"orders", "products", "logistics"},
		AuthType:    "oauth",
	},
	PlatformTikTok: {
		Code:        PlatformTikTok,
		Name:        "TikTok小店",
		Icon:        "tiktok",
		Description: "TikTok跨境电商",
		Features:    []string{"orders", "products"},
		AuthType:    "oauth",
	},
	PlatformJingqi: {
		Code:        PlatformJingqi,
		Name:        "京企直卖",
		Icon:        "jingqi",
		Description: "京东企业直卖平台",
		Features:    []string{"orders", "products"},
		AuthType:    "apikey",
	},

	// 后续扩展平台
	PlatformJD: {
		Code:        PlatformJD,
		Name:        "京东",
		Icon:        "jd",
		Description: "京东电商平台",
		Features:    []string{"orders", "products", "logistics"},
		AuthType:    "oauth",
	},
	PlatformXiaohongshu: {
		Code:        PlatformXiaohongshu,
		Name:        "小红书",
		Icon:        "xiaohongshu",
		Description: "小红书电商平台",
		Features:    []string{"orders", "products"},
		AuthType:    "oauth",
	},
	PlatformVip: {
		Code:        PlatformVip,
		Name:        "唯品会",
		Icon:        "vip",
		Description: "唯品会特卖平台",
		Features:    []string{"orders", "products"},
		AuthType:    "apikey",
	},
	Platform1688: {
		Code:        Platform1688,
		Name:        "1688",
		Icon:        "1688",
		Description: "阿里巴巴B2B批发平台",
		Features:    []string{"orders", "products"},
		AuthType:    "oauth",
	},
	PlatformWechat: {
		Code:        PlatformWechat,
		Name:        "微信小商店",
		Icon:        "wechat",
		Description: "微信小程序电商",
		Features:    []string{"orders", "products", "logistics"},
		AuthType:    "oauth",
	},
	PlatformOffline: {
		Code:        PlatformOffline,
		Name:        "实体店铺",
		Icon:        "shop",
		Description: "线下实体门店",
		Features:    []string{"orders", "manual"},
		AuthType:    "manual",
	},
	PlatformCustom: {
		Code:        PlatformCustom,
		Name:        "自定义平台",
		Icon:        "custom",
		Description: "第三方自定义平台",
		Features:    []string{"orders", "products", "webhook"},
		AuthType:    "apikey",
	},
}

// GetPlatformConfig 获取平台配置
func GetPlatformConfig(code PlatformType) (PlatformConfig, bool) {
	config, ok := PlatformConfigs[code]
	return config, ok
}

// GetAllPlatforms 获取所有平台配置
func GetAllPlatforms() []PlatformConfig {
	platforms := make([]PlatformConfig, 0, len(PlatformConfigs))
	for _, config := range PlatformConfigs {
		platforms = append(platforms, config)
	}
	return platforms
}
