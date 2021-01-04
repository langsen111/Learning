package main

// 合约基础数据
type ContractBasicData struct {
	TotalMembers                  int     `json:"totalMembers"`                  // 仲裁委员会人数上限
	SignUpTime                    int     `json:"signUpTime"`                    // 报名时间----天
	VoteTime                      int     `json:"voteTime"`                      // 投票时间----天
	SigningMargin                 int     `json:"signingMargin"`                 // 报名保证金--通证
	MinVoted                      float64 `json:"minVoted"`                      // 可加入委员会的最低得票率
	ArbitrationVoteNum            int     `json:"arbitrationVoteNum"`            // 委员会每个月可仲裁投票次数
	ArbitrationAppointedTime      int     `json:"arbitrationAppointedTime"`      // 委任时间--月
	ArbitrationAppointedSleepTime int     `json:"arbitrationAppointedSleepTime"` // 任期结束后的休息期---月
	Version                       int     `json:"version,omitempty"`             // 合约迭代版本号
	UpdateTime                    int64   `json:"updateTime,omitempty"`          // 更新时间
	CreateTime                    int64   `json:"createTime,omitempty"`          // 创建时间
	Tag                           string  `json:"tag,omitempty"`                 // 当前状态 0-最新使用的， 1-最新更新的，2-历史版本
}

// 报名信息
type SignUpInfo struct {
	BlanceNum             int             `json:"balanceNum"`            // 竞选剩余名额
	SignTime              int64           `json:"signTime"`              // 报名开始时间
	SignStopTime          int64           `json:"signStopTime"`          // 报名截止时间
	VoteTime              int64           `json:"voteTime"`              // 开始投票时间
	VoteStopTime          int64           `json:"voteStopTime"`          // 投票截止时间
	SignRounds            int             `json:"signRounds"`            // 报名轮次
	TotalSignNum          int             `json:"totalSignNum"`          // 总的投票成员人数
	VoteTotalSignNum      int             `json:"voteTotalSignNum"`      // 参与投票的总人数
	ApplicationList       []ApplicantInfo `json:"applicationList"`       // 总的报名人员信息列表
	ElectedPeople         []ApplicantInfo `json:"electedPeople"`         // 成功竞选人员列表
	LosingPeople          []ApplicantInfo `json:"losingPeople"`          // 竞选失败人员列表
	StatisticalVotingTime int64           `json:"statisticalVotingTime"` // 投票结算时间
}

// 报名人员信息
type ApplicantInfo struct {
	UserName     string  `json:"userName"`     // 报名人员用户名
	SignTime     int64   `json:"signTime"`     // 报名时间
	PledgeAmount int     `json:"pledgeAmount"` // 质押数量
	VoteNum      int     `json:"voteNum"`      // 被投票的次数
	VoteRate     float64 `json:"voteRate"`     // 得票率
}

// 仲裁委员会成员信息
type ArbitrationInfo struct {
	Username                 string      `json:"userName"`                 // 用户名
	StartSignTime            int64       `json:"time"`                     // 申请时间
	AppointmentStartTime     int64       `json:"appointmentStartTime"`     // 委任开始时间
	AppointmentStopTime      int64       `json:"appointmentStopTime"`      // 委任结束时间
	AppointmentStopSleepTime int64       `json:"appointmentStopSleepTime"` // 委任结束后休息时间
	Votenum                  int         `json:"voteNum"`                  // 参与投票的次数
	VoteEvents               []VoteEvent `json:"voteEvents"`               // 投票记录
}

// 仲裁委投票
type VoteEvent struct {
	VotedName  string `json:"votedName"`  // 被投票的人
	VotedEvent string `json:"votedEvent"` // 被投票的事件
	VoteNum    int    `json:"voteNum"`    // 投票的次数
	VoteTime   int64  `json:"voteTime"`   // 投票的时间
}

// 仲裁委员会组织
type ArbitrationMember struct {
	TotalNum     int               `json:"totalNum"`     // 仲裁委员会总的成员数
	RemainingNum int               `json:"remainingNum"` // 剩余仲裁委员会成员招募名额数
	Members      []ArbitrationInfo `json:members`        // 仲裁委成员
}

// 退出仲裁委员会组织休息期列表
type ArbitrationMemberSleep struct {
	TotalNum int               `json:"totalNum"` // 退出成员数
	Members  []ArbitrationInfo `json:members`    // 退出仲裁委员会组织休息期列表
}

const (
	TotalMembers                  = "totalMembers"
	SignUpTime                    = "signUpTime"
	VoteTime                      = "voteTime"
	SigningMargin                 = "signingMargin"
	MinVoted                      = "minVoted"
	ArbitrationVoteNum            = "arbitrationVoteNum"
	ArbitrationAppointedTime      = "arbitrationAppointedTime"
	ArbitrationAppointedSleepTime = "arbitrationAppointedSleepTime"
)

// 链码存储常量key
const (
	ARBITRATIONMEMBERS      = "Arbitration_Members"       // 仲裁委员会成员列表key
	ARBITRATIONMEMBERSSLEEP = "Arbitration_Members_Sleep" // 仲裁委员会成员退出休息期列表key
	CONTRACTBASICDATA       = "Contract_Basic_Data"       // 合约的基础数据key
	SIGNUPINFO              = "Sign_Up_Info"              // 报名信息key
	ALLVOTEDEVENT           = "All_Voted_Event"           // 所有人完成投票事件
)

// 链码存储类型识别
const (
	CONTRACTBASICDATA_TYPES      = iota // 合约基础数据类型
	SIGNUPINFO_TYPES                    // 报名数据
	APPLICANTINFO_TYPES                 // 申请报名
	ARBITRATIONINFO_TYPES               // 仲裁委员会成员信息
	VOTEEVENT_TYPES                     // 投票事件类型
	ARBITRATIONMEMBER_TYPES             // 仲裁委员会
	ARBITRATIONMEMBERSLEEP_TYPES        // 仲裁委员会退出休息
)

const (
	CurrentUseTag    = "0"
	CurrentUpdateTag = "1"
	HistoryTag       = "2"
)

// 合约初始化数值
var (
	Total_Members                            = 25   // 仲裁委员会人数上限
	Sign_Up_Time                             = 5    // 五个工作日的报名时间
	Vote_Time                                = 5    // 五个工作日的投票时间
	Signing_Margin                           = 100  // 报名保证金100通证
	Min_Voted                        float64 = 0.05 //  可加入委员会的最低得票率
	Arbitration_Vote_Num                     = 2    // 委员会每个月可仲裁投票次数
	Arbitration_Appointed_Time               = 12   // 委任时间12个月
	Arbitration_Appointed_Sleep_Time         = 3    // 任期结束后3个月的休息期
	Version                                  = 0    // 初始版本号为0
)


