package dto

// LeiGodLoginResp
// @description: 雷神登录返回
type LeiGodLoginResp struct {
	LoginInfo struct {
		AccountToken string `json:"account_token"` // Token
		ExpiryTime   string `json:"expiry_time"`   // 有效期
		NnToken      string `json:"nn_token"`
	} `json:"login_info"`
	UserInfo struct {
		Nickname   string `json:"nickname"`
		Email      string `json:"email"`
		Mobile     string `json:"mobile"`
		Avatar     string `json:"avatar"`
		RegionCode int    `json:"region_code"`
	} `json:"user_info"`
}

// LeiGodUserInfoResp
// @description: 雷神用户信息返回
type LeiGodUserInfoResp struct {
	UserPauseTime         int    `json:"user_pause_time"`
	Nickname              string `json:"nickname"`
	Email                 string `json:"email"`
	CountryCode           string `json:"country_code"`
	Mobile                string `json:"mobile"`
	UserName              string `json:"user_name"`
	MasterAccount         string `json:"master_account"`
	Birthday              string `json:"birthday"`
	PublicIp              string `json:"public_ip"`
	Sex                   string `json:"sex"`
	LastLoginTime         string `json:"last_login_time"`
	LastLoginIp           string `json:"last_login_ip"`
	PauseStatus           string `json:"pause_status"`    // 暂停状态
	PauseStatusId         int    `json:"pause_status_id"` // 暂停状态 0未暂停1已暂停
	LastPauseTime         string `json:"last_pause_time"` // 最后一次暂停时间
	VipLevel              string `json:"vip_level"`
	Avatar                string `json:"avatar"`
	AvatarNew             string `json:"avatar_new"`
	PackageId             string `json:"package_id"`
	IsSwitchPackage       int    `json:"is_switch_package"`
	PackageTitle          string `json:"package_title"`
	PackageLevel          string `json:"package_level"`
	BillingType           string `json:"billing_type"`
	Lang                  string `json:"lang"`
	StopedRemaining       string `json:"stoped_remaining"`
	ExpiryTime            string `json:"expiry_time"`      // 剩余时长
	ExpiryTimeSamp        int    `json:"expiry_time_samp"` // 剩余时长秒数
	Address               string `json:"address"`
	MobileContactType     string `json:"mobile_contact_type"`
	MobileContactNumber   string `json:"mobile_contact_number"`
	MobileContactTitle    string `json:"mobile_contact_title"`
	RegionCode            int    `json:"region_code"`
	IsPayUser             string `json:"is_pay_user"`
	WallLogSwitch         string `json:"wall_log_switch"`
	IsSetAdminPass        int    `json:"is_set_admin_pass"`
	ExpiredExperienceTime string `json:"expired_experience_time"`
	ExperienceExpiryTime  string `json:"experience_expiry_time"`
	ExperienceTime        int    `json:"experience_time"`
	FirstInvoiceDiscount  int    `json:"first_invoice_discount"`
	NnNumber              string `json:"nn_number"`
	UserSignature         string `json:"user_signature"`
	MobileExpiryTime      string `json:"mobile_expiry_time"`
	MobileExpiryTimeSamp  int    `json:"mobile_expiry_time_samp"`
	MobilePauseStatus     int    `json:"mobile_pause_status"`
	BlackExpiredTime      string `json:"black_expired_time"`
	MobileExperienceTime  string `json:"mobile_experience_time"`
	SuperTime             string `json:"super_time"`
	NowDate               string `json:"now_date"`
	NowTimeSamp           int    `json:"now_time_samp"`
	UserEarnMinutes       string `json:"user_earn_minutes"`
}
