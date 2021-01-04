package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"strconv"
	"time"
)

type VoteChaincode struct {
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	fn, args := stub.GetFunctionAndParameters()

	switch fn {
	case "init":
		return t.init(stub, args)
	case "updateContractData":
		return t.updateContractData(stub, args)
	case "getContractBasicInfo":
		return t.getContractBasicInfo(stub, args)
	case "application":
		return t.application(stub, args)
	case "voteArbitration":
		return t.voteArbitration(stub, args)
	case "statisticalVoting":
		return t.statisticalVoting(stub, args)
	case "exitArbitrationMembers":
		return t.exitArbitrationMembers(stub, args)
	case "getArbitrationMembers":
		return t.getArbitrationMembers(stub)
	case "getSignUpInfo":
		return t.getSignUpInfo(stub, args)
	case "getEixtArbitrationMembersSleep":
		return t.getEixtArbitrationMembersSleep(stub)
	case "detectionArbitrationMembers":
		return t.detectionArbitrationMembers(stub)
	case "detectionArbitrationMembersSleep":
		return t.detectionArbitrationMembersSleep(stub)
	default:
		return shim.Error("Invoke 调用方法有误！")
	}
}

// 初始化合约基础数据
func (this *VoteChaincode) init(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 是否需要验证调用者
	fmt.Println("==============function:init=================")

	// 判断是否已经初始化
	info, err := stub.GetState(genKey(CONTRACTBASICDATA, CurrentUseTag))
	fmt.Println(info)
	if err != nil || len(info) != 0 {
		return shim.Error("It's already initialized！")
	}

	if len(args) != 1 {
		return shim.Error("args length error")
	}
	fmt.Println(args[0])

	// 获取并格式化参数
	contractBasicDataArgs := args[0]
	contractBasicData := ContractBasicData{}
	if err = json.Unmarshal([]byte(contractBasicDataArgs), &contractBasicData); err != nil {
		return shim.Error("Get params error!")
	}

	currentTime := time.Now().Unix()

	// 合约基础数据初始化
	contractBasicData.Version = 1
	contractBasicData.UpdateTime = currentTime
	contractBasicData.CreateTime = currentTime
	contractBasicData.Tag = CurrentUseTag
	fmt.Printf("=====%#v====", contractBasicData)
	// 合约基础数据存储到链上
	if err = putDatatoChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), contractBasicData); err != nil {
		return shim.Error(err.Error())
	}

	// 初始化仲裁委员组织数据
	arbitrationMember := ArbitrationMember{
		TotalNum:     Total_Members,
		RemainingNum: Total_Members,
		Members:      []ArbitrationInfo{},
	}
	if err = putDatatoChainCode(stub, ARBITRATIONMEMBERS, arbitrationMember); err != nil {
		return shim.Error(err.Error())
	}

	// 初始化成功
	return shim.Success([]byte("Init data success."))
}

// 更新合约基础数据
func (this *VoteChaincode) updateContractData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=======update contract basic data....========")

	if len(args) != 1 {
		shim.Error("args length error")
	}

	// 获取并格式化参数
	updateDataArgs := args[0]
	updateDataMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(updateDataArgs), &updateDataMap); err != nil {
		return shim.Error("Get params error!")
	}

	// 获取当前使用合约基础数据
	contractBasicDatas, err := getDataFromChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), CONTRACTBASICDATA_TYPES)
	if err != nil {
		return shim.Error(err.Error())
	}
	contractBasicData := contractBasicDatas.(ContractBasicData)

	// 更改历史合约的tag,后续存储历史记录使用
	contractBasicData.Tag = HistoryTag
	contractBasicDataHistoryKey := genKey(genKey(CONTRACTBASICDATA, HistoryTag), strconv.Itoa(contractBasicData.Version))
	contractBasicDataHistory := contractBasicData


	// 更新合约基础数据
	for k, value := range updateDataMap {
		switch (k) {
		case TotalMembers:
			contractBasicData.TotalMembers = int(value.(float64))
		case SignUpTime:
			contractBasicData.SignUpTime = int(value.(float64))
		case VoteTime:
			contractBasicData.VoteTime = int(value.(float64))
		case SigningMargin:
			contractBasicData.SigningMargin = int(value.(float64))
		case MinVoted:
			contractBasicData.MinVoted = value.(float64)
		case ArbitrationVoteNum:
			contractBasicData.ArbitrationVoteNum = int(value.(float64))
		case ArbitrationAppointedTime:
			contractBasicData.ArbitrationAppointedTime = int(value.(float64))
		case ArbitrationAppointedSleepTime:
			contractBasicData.ArbitrationAppointedSleepTime = int(value.(float64))
		}
	}

	// 合约基础数据更新版本号自增
	contractBasicData.Version += 1
	contractBasicData.Tag = CurrentUseTag

	currentTime := time.Now().Unix()
	// 更新合约状态标记，判断是否是最新使用的合约
	// 获取当前轮次报名信息
	signUpInfoByte, err := stub.GetState(SIGNUPINFO)
	if err != nil {
		return shim.Error("====signUpInfoByte====get store error!")
	}
	if len(signUpInfoByte) != 0 {
		signUpInfo := SignUpInfo{}
		if err = json.Unmarshal(signUpInfoByte, &signUpInfo); err != nil {
			return shim.Error("json.Unmarshal(signUpInfoByte, &signUpInfo) ======error")
		}
		// 在报名期内更新了合约
		if signUpInfo.SignTime <= currentTime && currentTime <= signUpInfo.SignStopTime {
			fmt.Println(".....在报名期内更新了合约......\n")
			contractBasicData.Tag = CurrentUpdateTag

			// 判断是否存在更新中的合约 有的话直接将状态改为历史
			updatesContractBasicDataByte, err := stub.GetState(genKey(CONTRACTBASICDATA, CurrentUpdateTag))
			if err != nil {
				return shim.Error("stub.GetState(genKey(CONTRACTBASICDATA, CurrentUpdateTag)) ==== error")
			}
			if len(updatesContractBasicDataByte) != 0 {
				updateContractBasicData := ContractBasicData{}
				if err = json.Unmarshal(updatesContractBasicDataByte, &updateContractBasicData); err != nil {
					return shim.Error("update data failure==========...")
				}
				// 更新历史版本
				updateContractBasicData.Tag = HistoryTag
				if err = putDatatoChainCode(stub, genKey(genKey(CONTRACTBASICDATA, HistoryTag), strconv.Itoa(updateContractBasicData.Version)), updateContractBasicData); err != nil {
					return shim.Error(err.Error())
				}
				contractBasicData.Version = updateContractBasicData.Version + 1
			}

			contractBasicData.UpdateTime = currentTime

			// 将更新的数据存储到链上
			if err = putDatatoChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUpdateTag), contractBasicData); err != nil {
				shim.Error(fmt.Sprintf("updateContractBasicDataByte ==== put store error! err: %v", err.Error()))
			}
			return shim.Success([]byte("Update current to update success."))
		}
	}

	// 存储合约历史版本
	fmt.Println(contractBasicDataHistoryKey)
	if err = putDatatoChainCode(stub, contractBasicDataHistoryKey, contractBasicDataHistory); err != nil {
		return shim.Error(err.Error())
	}
	contractBasicData.UpdateTime = currentTime

	// 将最新使用的合约信息存储到链上
	if err = putDatatoChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), contractBasicData); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(" Update data success."))
}

// 获取合约基础数据  参数小与0表示获取当前合约信息，等于0表示即将生效的合约信息，参数大于0表示历史版本的合约信息
func (t *VoteChaincode) getContractBasicInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=====function:getContractBasicInfo====")
	if len(args) != 1 {
		shim.Error("args length error")
	}
	fmt.Println(args[0])
	version, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("getContractBasicInfo == Input params error,%v", err.Error()))
	}

	var contractBasicDataByte []byte
	if version < 0 {
		// 获取最新合约数据信息
		contractBasicDataByte, err = stub.GetState(genKey(CONTRACTBASICDATA, CurrentUseTag))
	} else if version > 0 {
		// 获取历史版本的合约信息
		fmt.Println(genKey(genKey(CONTRACTBASICDATA, HistoryTag), args[0]))
		contractBasicDataByte, err = stub.GetState(genKey(genKey(CONTRACTBASICDATA, HistoryTag), args[0]))
	} else {
		contractBasicDataByte, err = stub.GetState(genKey(CONTRACTBASICDATA, CurrentUpdateTag))
	}

	if err != nil || len(contractBasicDataByte) == 0 {
		fmt.Println(err.Error())
		return shim.Error(fmt.Sprintf("Params error || have nothing info, err: %v", err.Error()))
	}
	return shim.Success(contractBasicDataByte)
}

// 报名加入仲裁委员会
func (t *VoteChaincode) application(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("==============function:application=================")

	if len(args) != 3 {
		return shim.Error("Args length error!")
	}

	// 获取参数
	totalSignNum, err := strconv.Atoi(args[0])
	userName := args[1]
	pledgeAmount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("args[0] error!")
	}

	// 判断是否在退出休息期列表中
	_, err = isInArbitrationMemberSleep(stub, []string{userName})
	if err != nil {
		return shim.Error(err.Error())
	}
	// 判断是否已经在仲裁委列表中
	_, err = isInArbitrationMember(stub, userName)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 获取当前使用合约基础数据
	contractBasicDatas, err := getDataFromChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), CONTRACTBASICDATA_TYPES)
	if err != nil {
		return shim.Error(err.Error())
	}
	contractBasicData := contractBasicDatas.(ContractBasicData)

	// 获取报名信息
	signUpInfoByte, err := stub.GetState(SIGNUPINFO)
	if err != nil {
		return shim.Error("get store error!")
	}

	nowTime := time.Now()
	// 报名人员信息
	applicationInfo := ApplicantInfo{
		UserName:     userName,
		SignTime:     nowTime.Unix(),
		PledgeAmount: pledgeAmount,
		VoteNum:      0,
		VoteRate:     0,
	}

	signUpInfo := SignUpInfo{}
	// 判断是否是最新一轮的第一个开始报名
	if len(signUpInfoByte) == 0 {
		// 处理第一个开始报名的信息
		signUpInfo = SignUpInfo{
			BlanceNum: contractBasicData.TotalMembers,
			SignTime:  nowTime.Unix(),
			//SignStopTime:     nowTime.AddDate(0, 0, contractBasicData.SignUpTime).Unix(),
			//VoteTime:         nowTime.AddDate(0, 0, contractBasicData.SignUpTime).Unix(),
			//VoteStopTime:     nowTime.AddDate(0, 0, contractBasicData.SignUpTime+contractBasicData.VoteTime).Unix(),

			SignStopTime: nowTime.Unix() + int64(contractBasicData.SignUpTime*60),
			VoteTime:     nowTime.Unix() + int64(contractBasicData.SignUpTime*60),
			VoteStopTime: nowTime.Unix() + int64((contractBasicData.SignUpTime+contractBasicData.VoteTime)*60),

			SignRounds:            1,
			TotalSignNum:          totalSignNum,
			VoteTotalSignNum:      0,
			ApplicationList:       []ApplicantInfo{applicationInfo},
			ElectedPeople:         []ApplicantInfo{},
			LosingPeople:          []ApplicantInfo{},
			StatisticalVotingTime: 0,
		}
	} else { // 处理非第一个开始报名的人员信息
		if err = json.Unmarshal(signUpInfoByte, &signUpInfo); err != nil {
			return shim.Error("SignUpInfoByte json.Unmarshal failure!...")
		}
		// 处理新一轮报名事件
		if signUpInfo.BlanceNum > 0 && nowTime.Unix() >= signUpInfo.VoteStopTime {
			// 获取仲裁委员会信息
			arbitrationMembers, err := getDataFromChainCode(stub, ARBITRATIONMEMBERS, ARBITRATIONMEMBER_TYPES)
			if err != nil{
				return shim.Error(err.Error())
			}
			arbitrationMember := arbitrationMembers.(ArbitrationMember)
			// 处理第一个开始报名的信息
			signUpInfo = SignUpInfo{
				BlanceNum: arbitrationMember.RemainingNum,
				SignTime:  nowTime.Unix(),
				//SignStopTime:     nowTime.AddDate(0, 0, contractBasicData.SignUpTime).Unix(),
				//VoteTime:         nowTime.AddDate(0, 0, contractBasicData.SignUpTime).Unix(),
				//VoteStopTime:     nowTime.AddDate(0, 0, contractBasicData.SignUpTime+contractBasicData.VoteTime).Unix(),
				SignStopTime: nowTime.Unix() + int64(contractBasicData.SignUpTime*60),
				VoteTime:     nowTime.Unix() + int64(contractBasicData.SignUpTime*60),
				VoteStopTime: nowTime.Unix() + int64((contractBasicData.SignUpTime+contractBasicData.VoteTime)*60),

				SignRounds:            signUpInfo.SignRounds + 1,
				TotalSignNum:          totalSignNum,
				VoteTotalSignNum:      0,
				ApplicationList:       []ApplicantInfo{applicationInfo},
				ElectedPeople:         []ApplicantInfo{},
				LosingPeople:          []ApplicantInfo{},
				StatisticalVotingTime: 0,
			}
		} else if signUpInfo.SignTime <= nowTime.Unix() && nowTime.Unix() <= signUpInfo.SignStopTime { // 判断是否在报名期内
			// 判断是否重复报名
			for _, v := range signUpInfo.ApplicationList {
				if userName == v.UserName {
					return shim.Error("Repeat registration, please adjust！")
				}
			}
			// 加入到竞选列表中
			signUpInfo.ApplicationList = append(signUpInfo.ApplicationList, applicationInfo)
		} else { // 不在规定期内报名
			return shim.Error("Not during the registration period！")
		}
	}
	// 存储到链上
	if err = putDatatoChainCode(stub, SIGNUPINFO, signUpInfo); err != nil {
		return shim.Error("signUpInfoJsonByte put store failure!...")
	}
	return shim.Success([]byte("registration successful!!"))
}

// 参选者被投票
func (t *VoteChaincode) voteArbitration(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("===========voteArbitration============")
	if len(args) != 1 {
		return shim.Error("args length error")
	}
	userName := args[0]

	// 获取报名信息
	signUpInfos, err := getDataFromChainCode(stub, SIGNUPINFO, SIGNUPINFO_TYPES)
	if err != nil{
		return shim.Error(err.Error())
	}
	signUpInfo := signUpInfos.(SignUpInfo)

	currentTime := time.Now()
	// 判断是否所有人都参与投票了
	if signUpInfo.VoteTotalSignNum == signUpInfo.TotalSignNum {
		return shim.Error("Everyone has voted. The polls are closed！...")
	}

	// 获取报名列表
	applicationList := signUpInfo.ApplicationList
	if len(applicationList) < 1 {
		return shim.Error("Voting error！...")
	}

	var flag bool
	// 判断是否在投票期内
	if signUpInfo.VoteTime <= currentTime.Unix() && currentTime.Unix() <= signUpInfo.VoteStopTime {
		// 判断投票的对象是否在申请列表中
		flag = false
		for i := 0; i < len(applicationList); i++ {
			if applicationList[i].UserName == userName {
				// 参选人票数+1
				applicationList[i].VoteNum += 1
				applicationList[i].VoteRate, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(applicationList[i].VoteNum)/float64(signUpInfo.TotalSignNum)), 64)
				fmt.Printf("==========投票内容======%#v====", applicationList[i])

				// 总的参与投票数+1
				signUpInfo.VoteTotalSignNum += 1
				flag = true
			}
		}
		if flag == false {
			return shim.Error("The people who voted are not on the list！...")
		}

	} else {
		return shim.Error("Out of voting period！...")
	}

	// 所有人完成投票设置监听事件
	if signUpInfo.VoteTotalSignNum == signUpInfo.TotalSignNum {
		if err = stub.SetEvent(ALLVOTEDEVENT, []byte(ALLVOTEDEVENT)); err != nil {
			shim.Error("set event error!..." + err.Error())
			fmt.Println("set event error!..." + err.Error())
		}
	}

	// 将更新完的数据存入链上
	if err = putDatatoChainCode(stub, SIGNUPINFO, signUpInfo); err != nil{
		return shim.Error("Put store error!...")
	}
	return shim.Success([]byte("vote successful..."))
}

// 统计投票结果
func (t *VoteChaincode) statisticalVoting(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========StatisticalVoting===========")

	// 获取当前使用合约基础数据
	contractBasicDatas, err := getDataFromChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), CONTRACTBASICDATA_TYPES)
	if err != nil {
		return shim.Error(err.Error())
	}
	contractBasicData := contractBasicDatas.(ContractBasicData)

	// 获取报名信息
	signUpInfos, err := getDataFromChainCode(stub, SIGNUPINFO, SIGNUPINFO_TYPES)
	if err != nil{
		return shim.Error(err.Error())
	}
	signUpInfo := signUpInfos.(SignUpInfo)

	// 判断是否已经统计过了
	fmt.Println(signUpInfo.StatisticalVotingTime)
	if signUpInfo.StatisticalVotingTime > 0 {
		return shim.Error("\n The poll is over, please check the history!")
	}

	// 获取投票最低通过率
	minVotedRate := contractBasicData.MinVoted
	// 获取竞选名额
	blanceNum := signUpInfo.BlanceNum
	// 获取报名投票轮次
	signRouds := signUpInfo.SignRounds
	if blanceNum < 1 {
		return shim.Error("There are no electoral seats left！...")
	}

	// 获取仲裁委员会组织
	arbitrationMembers, err := getDataFromChainCode(stub, ARBITRATIONMEMBERS, ARBITRATIONMEMBER_TYPES)
	if err != nil{
		return shim.Error(err.Error())
	}
	arbitrationMember := arbitrationMembers.(ArbitrationMember)

	currentTime := time.Now()
	// 统计结果的筛选条件
	if signUpInfo.TotalSignNum == signUpInfo.VoteTotalSignNum || currentTime.Unix() >= signUpInfo.VoteStopTime {
		// 获取报名列表
		applicationList := signUpInfo.ApplicationList
		if len(applicationList) < 1 {
			return shim.Error("StatisticalVoting error！...")
		}
		lenApplicationList := len(applicationList)
		// 对报名人员的票数进行降序排序,若票数相等则比较报名时间，靠前的排在高位
		fmt.Println("排序开始============\n")
		for i := 0; i < lenApplicationList-1; i++ {
			for j := i + 1; j < lenApplicationList; j++ {
				if applicationList[i].VoteRate < applicationList[j].VoteRate {
					applicationList[i], applicationList[j] = applicationList[j], applicationList[i]
				}else if applicationList[i].VoteRate == applicationList[j].VoteRate{
					// 判断报名时间
					if applicationList[i].SignTime > applicationList[j].SignTime{
						applicationList[i], applicationList[j] = applicationList[j], applicationList[i]
					}
				}
			}
		}

		fmt.Printf("排序结果打印==========%#v===\n", applicationList)
		// 晒选出满足条件的人员
		for i := 0; i < lenApplicationList; i++ {
			fmt.Println("进入筛选条件======\n")
			// 满足竞选成功的加入到竞选成功列表,竞选失败的加入到失败列表
			if applicationList[i].VoteRate >= minVotedRate && blanceNum > 0 {
				signUpInfo.ElectedPeople = append(signUpInfo.ElectedPeople, applicationList[i])
				// 加入到仲裁委员会
				arbitrationMember.Members = append(arbitrationMember.Members, ArbitrationInfo{
					Username:             applicationList[i].UserName,
					StartSignTime:        applicationList[i].SignTime,
					AppointmentStartTime: currentTime.Unix(),
					//AppointmentStopTime:  currentTime.AddDate(0, contractBasicData.ArbitrationAppointedTime, 0).Unix(),
					AppointmentStopTime: currentTime.Unix() + int64(contractBasicData.ArbitrationAppointedTime*60),
					VoteEvents:          []VoteEvent{},
					Votenum:             0,
				})
				blanceNum -= 1
			} else {
				signUpInfo.LosingPeople = append(signUpInfo.LosingPeople, applicationList[i])
			}
		}
		// 更新最后剩余的名额
		signUpInfo.BlanceNum = blanceNum
		arbitrationMember.RemainingNum = blanceNum

		// 更新投票结果汇总时间
		signUpInfo.StatisticalVotingTime = currentTime.Unix()
	} else {
		return shim.Error("Conditions are not met. Please try again later！...")
	}

	// 转为json对象
	signUpInfoBytes, err := json.Marshal(signUpInfo)
	if err != nil {
		return shim.Error("signUpInfoBytes=====json marshal error!...")
	}
	// 将这轮的报名投票信息记录为历史记录，结束报名或者等待下一轮报名的开启
	if err = putDatatoChainCode(stub, genKey(SIGNUPINFO, strconv.Itoa(signRouds)), signUpInfo); err != nil {
		return shim.Error("signUpInfoBytes====put store error!...")
	}
	if err = putDatatoChainCode(stub, SIGNUPINFO, signUpInfo); err != nil {
		return shim.Error("signUpInfoBytes====put store error!...")
	}

	// 存储仲裁委员信息
	if err = putDatatoChainCode(stub, ARBITRATIONMEMBERS, arbitrationMember); err != nil {
		return shim.Error("arbitrationMemberBytes === put store error!...")
	}

	// 更新合约基础数据
	if err = replaceContractInfo(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(signUpInfoBytes)
}

// 获取当前报名信息，或者获取历史报名信息
func (t *VoteChaincode) getSignUpInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("==============function:getSignUpInfo=================")

	lenArgs := len(args)
	var signUpInfoByte []byte
	var err error
	switch lenArgs {
	case 0:
		// 获取报名信息
		signUpInfoByte, err = stub.GetState(SIGNUPINFO)
	case 1:
		// 获取历史轮次报名信息
		signUpInfoByte, err = stub.GetState(genKey(SIGNUPINFO, args[0]))

	default:
		return shim.Error("args length error!...")
	}
	if err != nil || signUpInfoByte == nil {
		return shim.Error("getSignUpInfo error or don't search sign up info!...")
	}
	return shim.Success(signUpInfoByte)
}

// 退出仲裁委员会
func (t *VoteChaincode) exitArbitrationMembers(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("==============function:exitArbitrationMembers=================")

	//// 获取调用者身份
	//from := initiator(stub)
	var userName string
	var isForcedExit string
	switch len(args) {
	case 1:
		userName = args[0]
	case 2:
		userName = args[0]
		isForcedExit = args[1]
	default:
		return shim.Error("args length error!...")
	}

	if err := exitArbitrationMembers(stub, []string{userName}, isForcedExit); err != nil {
		return shim.Error(fmt.Sprintf("====username:{%v} ======exit error!...err:%v", userName, err))
	}

	return shim.Success([]byte("success!..."))
}

// 获取仲裁委员会成员列表
func (t *VoteChaincode) getArbitrationMembers(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("==============function:getArbitrationMembers=================")

	ArbitrationMemberBytes, err := stub.GetState(ARBITRATIONMEMBERS)
	if err != nil {
		shim.Error("Failed to obtain information about members of the arbitration Commission！...")
	}
	if ArbitrationMemberBytes == nil {
		shim.Success([]byte("No member has yet joined！..."))
	}
	return shim.Success(ArbitrationMemberBytes)
}

// 获取休息期内的委员会成员
func (t *VoteChaincode) getEixtArbitrationMembersSleep(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("=========function:getEixtArbitrationMembersSleep======\n")
	// 获取退出仲裁委休息期数据
	arbitrationMemberSleepByte, err := stub.GetState(ARBITRATIONMEMBERSSLEEP)
	if err != nil {
		return shim.Error("get ARBITRATIONMEMBERSSLEEP error! " + err.Error())
	}
	if arbitrationMemberSleepByte == nil {
		return shim.Error("have nothing...")
	}
	return shim.Success(arbitrationMemberSleepByte)
}

// 检测仲裁委员成员委任时间是否到期,如果到期则退出仲裁委列表
func (t *VoteChaincode) detectionArbitrationMembers(stub shim.ChaincodeStubInterface) (pb.Response) {
	// 获取仲裁委员会成员
	arbitrationMembers, err := getDataFromChainCode(stub, ARBITRATIONMEMBERS, ARBITRATIONMEMBER_TYPES)
	if err != nil{
		return shim.Error(err.Error())
	}
	arbitrationMember := arbitrationMembers.(ArbitrationMember)

	fmt.Printf("member========%#v============长度：%v \n", arbitrationMember.Members, len(arbitrationMember.Members))
	memberLength := len(arbitrationMember.Members)
	if memberLength < 1 {
		return shim.Error("have nothing members!")
	}
	currentTime := time.Now().Unix()
	var userName []string
	for i := 0; i < memberLength; i++ {
		fmt.Println("=================:{}=", i)
		// 如果达到委任时间
		if currentTime >= arbitrationMember.Members[i].AppointmentStopTime {
			fmt.Printf("userNames:%#v", arbitrationMember.Members[i].Username)
			userName = append(userName, arbitrationMember.Members[i].Username)
		}
	}
	if len(userName) != 0 {
		if err = exitArbitrationMembers(stub, userName, ""); err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success([]byte("success!"))
}

// 清除到期休息期的信息
func (t *VoteChaincode) detectionArbitrationMembersSleep(stub shim.ChaincodeStubInterface) (pb.Response) {

	// 获取链码上数据
	arbitrationMemberSleeps, err := getDataFromChainCode(stub, ARBITRATIONMEMBERSSLEEP, ARBITRATIONMEMBERSLEEP_TYPES)
	if err != nil{
		return shim.Error(err.Error())
	}
	arbitrationMemberSleep := arbitrationMemberSleeps.(ArbitrationMemberSleep)

	if len(arbitrationMemberSleep.Members) == 0{
		return shim.Error("don't have Members...")
	}
	currentTime := time.Now().Unix()
	var userName []string
	for _, v := range arbitrationMemberSleep.Members {
		if currentTime >= v.AppointmentStopSleepTime {
			userName = append(userName, v.Username)
		}
	}

	if len(userName) != 0 {
		_, err = isInArbitrationMemberSleep(stub, userName)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success([]byte("Operation is successful！"))
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Println("vote chaincode start err")
	}
}

