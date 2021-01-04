package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"strconv"
	"strings"
	"time"
)

// 从链码上根据key获取数据
func getDataFromChainCode(stub shim.ChaincodeStubInterface, key string, types int) (interface{}, error) {

	dataInfoByte, err := stub.GetState(key)
	if err != nil || len(dataInfoByte) == 0 {
		fmt.Println(err.Error())
		return nil, errors.New(fmt.Sprintf("get info error!"))
	}
	switch types {
	case CONTRACTBASICDATA_TYPES:
		dataInfoFormat := ContractBasicData{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case SIGNUPINFO_TYPES:
		dataInfoFormat := SignUpInfo{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case APPLICANTINFO_TYPES:
		dataInfoFormat := ApplicantInfo{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case ARBITRATIONINFO_TYPES:
		dataInfoFormat := ArbitrationInfo{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case VOTEEVENT_TYPES:
		dataInfoFormat := VoteEvent{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case ARBITRATIONMEMBER_TYPES:
		dataInfoFormat := ArbitrationMember{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	case ARBITRATIONMEMBERSLEEP_TYPES:
		dataInfoFormat := ArbitrationMemberSleep{}
		if err = json.Unmarshal(dataInfoByte, &dataInfoFormat); err != nil {
			return nil, errors.New(fmt.Sprintf(" contractBasicDataByte json unmarshal error!"))
		}
		return dataInfoFormat, nil
	default:
		return nil, errors.New("types error!...")
	}
}

// 将数据存储到链码上
func putDatatoChainCode(stub shim.ChaincodeStubInterface, key string, value interface{}) (error) {
	dataInfoBytes, err := json.Marshal(value)
	if err != nil {
		return errors.New("Json marshal dataInfoBytes error!...")
	}
	if err = stub.PutState(key, dataInfoBytes); err != nil {
		return errors.New("put dataInfoBytes === error!...")
	}
	return nil
}

// 退出仲裁委员会
func exitArbitrationMembers(stub shim.ChaincodeStubInterface, userName []string, isForcedExit string) (error) {

	// 获取退出仲裁委休息期数据
	arbitrationMemberSleepByte, err := stub.GetState(ARBITRATIONMEMBERSSLEEP)
	if err != nil {
		return errors.New("get ARBITRATIONMEMBERSSLEEP error! " + err.Error())
	}
	arbitrationMemberSleep := ArbitrationMemberSleep{}
	if arbitrationMemberSleepByte == nil {
		arbitrationMemberSleep = ArbitrationMemberSleep{
			TotalNum: 0,
			Members:  []ArbitrationInfo{},
		}
	} else {
		if err = json.Unmarshal(arbitrationMemberSleepByte, &arbitrationMemberSleep); err != nil {
			return errors.New("arbitrationMemberSleepByte------json marshal error!...")
		}
	}

	currentTime := time.Now()
	// 获取仲裁委员会成员
	arbitrationMembers, err := getDataFromChainCode(stub, ARBITRATIONMEMBERS, ARBITRATIONMEMBER_TYPES)
	if err != nil{
		return errors.New(err.Error())
	}
	arbitrationMember := arbitrationMembers.(ArbitrationMember)

	fmt.Printf("=======Members :%#v", arbitrationMember.Members)
	if len(arbitrationMember.Members) == 0 {
		return errors.New("Username is not exists!...")
	}
	// 检验调用者是否在仲裁委成员当中
	isExist := false
reback:
	for k, v := range arbitrationMember.Members {
		for i := 0; i < len(userName); i++ {
			fmt.Printf("params======%v, 程序的=====%v  \n", userName, v.Username)
			if userName[i] == v.Username {
				isExist = true
				userName = append(userName[:i], userName[i+1:]...)
				// 判断是否强制退出
				if isForcedExit == "" {
					// 判断是否达到委任时间退出
					if currentTime.Unix() >= v.AppointmentStopTime {
						goto updateData
					} else {
						return errors.New("The time of appointment has not been met and there is no way to withdraw at present！...")
					}
				} else { // 强制退出
					v.AppointmentStopTime = currentTime.Unix()
					goto updateData
				}
				// 增加更新数据标签
			updateData:
				// 获取当前使用合约基础数据
				contractBasicDatas, err := getDataFromChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), CONTRACTBASICDATA_TYPES)
				if err != nil {
					return errors.New(err.Error())
				}
				contractBasicData := contractBasicDatas.(ContractBasicData)
				// 更新退出委员会信息
				arbitrationMemberSleep.TotalNum += 1
				//v.AppointmentStopSleepTime = currentTime.AddDate(0, contractBasicData.ArbitrationAppointedSleepTime, 0).Unix()
				v.AppointmentStopSleepTime = v.AppointmentStopTime + int64(contractBasicData.ArbitrationAppointedSleepTime*60)

				arbitrationMemberSleep.Members = append(arbitrationMemberSleep.Members, v)

				// 更新仲裁委员会人员信息
				arbitrationMember.RemainingNum += 1
				arbitrationMember.Members = append(arbitrationMember.Members[:k], arbitrationMember.Members[k+1:]...)
				goto reback
			}
		}
	}
	if isExist == false {
		return errors.New("The member is not in the list!...")
	}

	// 仲裁委信息存储到链上
	if err = putDatatoChainCode(stub, ARBITRATIONMEMBERS, arbitrationMember); err != nil {
		return errors.New("==ARBITRATIONMEMBERS-----put store error!...")
	}
	if err = putDatatoChainCode(stub, ARBITRATIONMEMBERSSLEEP, arbitrationMemberSleep); err != nil {
		return errors.New("==ARBITRATIONMEMBERSSLEEP-----put store error!...")
	}
	return nil
}

// 判断是否在仲裁委员会中
func isInArbitrationMember(stub shim.ChaincodeStubInterface, userName string) (bool, error) {
	fmt.Println("======isInArbitrationMember========")

	arbitrationMemberByte, err := stub.GetState(ARBITRATIONMEMBERS)
	if err != nil {
		return false, errors.New("get ARBITRATIONMEMBERS error! " + err.Error())
	}
	currentTime := time.Now()
	if len(arbitrationMemberByte) != 0 {
		arbitrationMember := ArbitrationMember{}
		err = json.Unmarshal(arbitrationMemberByte, &arbitrationMember)
		if err != nil {
			return false, errors.New(err.Error())
		}

		for _, v := range arbitrationMember.Members {
			if v.Username == userName {
				// 判断是否到了委任时间
				if currentTime.Unix() > v.AppointmentStopTime {
					if err = exitArbitrationMembers(stub, []string{userName}, ""); err != nil {
						return false, errors.New("...exit error!...")
					}
				} else {
					return false, errors.New("Already a member of the committee, no need to sign up again！...")
				}
			}
		}
	}
	return true, nil
}

// 判断是否在退出休息期中,true表示不在休息期内
func isInArbitrationMemberSleep(stub shim.ChaincodeStubInterface, userName []string) (bool, error) {
	fmt.Println("======isInArbitrationMemberSleep========")

	arbitrationMemberSleepByte, err := stub.GetState(ARBITRATIONMEMBERSSLEEP)
	if err != nil {
		return false, errors.New("get ARBITRATIONMEMBERSSLEEP error! " + err.Error())
	}
	if len(arbitrationMemberSleepByte) != 0 {
		arbitrationMemberSleep := ArbitrationMemberSleep{}
		err = json.Unmarshal(arbitrationMemberSleepByte, &arbitrationMemberSleep)
		if err != nil {
			return false, errors.New(err.Error())
		}
		currentTime := time.Now()
		flag := false
	reback:
		for k, v := range arbitrationMemberSleep.Members {
			for i := 0; i < len(userName); i++ {
				// 判断是否在列表中
				if v.Username == userName[i] {
					// 判断是否过了休息期,过了休息期则将从休息期列表中清除
					if currentTime.Unix() >= v.AppointmentStopSleepTime {
						flag = true
						userName = append(userName[:i], userName[i+1:]...)
						arbitrationMemberSleep.Members = append(arbitrationMemberSleep.Members[:k], arbitrationMemberSleep.Members[k+1:]...)
						arbitrationMemberSleep.TotalNum -= 1
						goto reback
					} else {
						return false, errors.New("Please confirm the time to apply for registration！...")
					}
				}

			}
		}
		if flag == true {
			arbitrationMemberSleepBytes, err := json.Marshal(arbitrationMemberSleep)
			if err != nil {
				return false, err
			}
			// 更新链上数据
			if err = stub.PutState(ARBITRATIONMEMBERSSLEEP, arbitrationMemberSleepBytes); err != nil {
				return false, errors.New("==ARBITRATIONMEMBERSSLEEP-----put store error!...")
			}
		}
	}
	return true, nil
}

// 将合约更新数据更新为最新使用的数据
func replaceContractInfo(stub shim.ChaincodeStubInterface) error {
	// 判断是否存在更新中的合约 有的话直接将状态改为历史
	updatesContractBasicDataByte, err := stub.GetState(genKey(CONTRACTBASICDATA, CurrentUpdateTag))
	if err != nil {
		return errors.New("stub.GetState(genKey(CONTRACTBASICDATA, CurrentUpdateTag)) ==== error")
	}
	if len(updatesContractBasicDataByte) != 0 {
		updateContractBasicData := ContractBasicData{}
		if err = json.Unmarshal(updatesContractBasicDataByte, &updateContractBasicData); err != nil {
			return errors.New("update data failure==========...")
		}
		// 更新为最新使用的版本
		updateContractBasicData.Tag = CurrentUseTag

		// 获取当前使用合约基础数据
		contractBasicDatas, err := getDataFromChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), CONTRACTBASICDATA_TYPES)
		if err != nil {
			return errors.New(err.Error())
		}
		contractBasicData := contractBasicDatas.(ContractBasicData)

		contractBasicData.Tag = HistoryTag
		// 存储为最新使用版本
		if err = putDatatoChainCode(stub, genKey(CONTRACTBASICDATA, CurrentUseTag), updateContractBasicData); err != nil {
			return errors.New(fmt.Sprintf("===updatesContractBasicDataByte===old contract info put store failre! err :%v", err.Error()))
		}
		// 存储到历史版本
		if err = putDatatoChainCode(stub, genKey(genKey(CONTRACTBASICDATA, HistoryTag), strconv.Itoa(contractBasicData.Version)), contractBasicData); err != nil {
			return errors.New(fmt.Sprintf("===updatesContractBasicDataByte===old contract info put store failre! err :%v", err.Error()))
		}
		// 删除update数据
		if err = stub.DelState(genKey(CONTRACTBASICDATA, CurrentUseTag)); err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

// 根据username和event生成key
func genKey(username, event string) string {
	var info []string = []string{username, event}
	return strings.Join(info, "_")
}
