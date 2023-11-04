package business_service

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business-repository/business_common"
	"github.com/zapscloud/golib-business-repository/business_repository"
	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// UserTypeService - Accounts Service structure
type UserTypeService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(userTypeId string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(userTypeId string, indata utils.Map) (utils.Map, error)
	Delete(userTypeId string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// userTypeBaseService - Accounts Service structure
type userTypeBaseService struct {
	db_utils.DatabaseService
	dbRegion            db_utils.DatabaseService
	daoUserType         business_repository.UserTypeDao
	daoPlatformBusiness platform_repository.BusinessDao
	child               UserTypeService
	businessID          string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewUserTypeService(props utils.Map) (UserTypeService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("UserTypeService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := userTypeBaseService{}

	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_services.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessID = businessId

	// Instantiate other services
	p.daoUserType = business_repository.NewUserTypeDao(p.dbRegion.GetClient(), businessId)
	p.daoPlatformBusiness = platform_repository.NewBusinessDao(p.GetClient())

	_, err = p.daoPlatformBusiness.Get(p.businessID)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid business id",
			ErrorDetail: "Given business id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, nil
}

func (p *userTypeBaseService) EndService() {
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

// List - List All records
func (p *userTypeBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("AccountService::FindAll - Begin")

	daoUserType := p.daoUserType
	response, err := daoUserType.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("AccountService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *userTypeBaseService) Get(userTypeId string) (utils.Map, error) {
	log.Printf("AccountService::FindByCode::  Begin %v", userTypeId)

	data, err := p.daoUserType.Get(userTypeId)
	log.Println("AccountService::FindByCode:: End ", err)
	return data, err
}

func (p *userTypeBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("AccountService::FindByCode::  Begin ", filter)

	data, err := p.daoUserType.Find(filter)
	log.Println("AccountService::FindByCode:: End ", data, err)
	return data, err
}

func (p *userTypeBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[business_common.FLD_USERTYPE_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("stftyp")
		log.Println("Unique Account ID", uid)
		indata[business_common.FLD_USERTYPE_ID] = uid
		dataval = indata[business_common.FLD_USERTYPE_ID]
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Account ID:", dataval)

	_, err := p.daoUserType.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Account ID !", ErrorDetail: "Given Account ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoUserType.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *userTypeBaseService) Update(userTypeId string, indata utils.Map) (utils.Map, error) {

	log.Println("AccountService::Update - Begin")

	data, err := p.daoUserType.Get(userTypeId)
	if err != nil {
		return data, err
	}

	data, err = p.daoUserType.Update(userTypeId, indata)
	log.Println("AccountService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *userTypeBaseService) Delete(userTypeId string, delete_permanent bool) error {

	log.Println("AccountService::Delete - Begin", userTypeId)

	daoUserType := p.daoUserType
	if delete_permanent {
		result, err := daoUserType.Delete(userTypeId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(userTypeId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("UserTypeService::Delete - End")
	return nil
}

func (p *userTypeBaseService) errorReturn(err error) (UserTypeService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
