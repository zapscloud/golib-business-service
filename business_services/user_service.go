package business_services

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business/business_common"
	"github.com/zapscloud/golib-business/business_repository"
	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// UserService - Users Service structure
type UserService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(userid string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(dataUser utils.Map) (utils.Map, error)
	Update(userid string, indata utils.Map) (utils.Map, error)
	Delete(userid string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// userBaseService - Users Service structure
type userBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoUser     business_repository.UserDao
	daoBusiness platform_repository.BusinessDao
	child       UserService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewUserService(props utils.Map) (UserService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"
	log.Printf("UserService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := userBaseService{}
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
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid business_id",
			ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// UserBaseService - Close all the services
func (p *userBaseService) EndService() {
	log.Printf("EndUserService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *userBaseService) initializeService() {
	log.Printf("UserService:: initializeService ")
	p.daoUser = business_repository.NewUserDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *userBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("UserService::FindAll - Begin")

	daoUser := p.daoUser
	response, err := daoUser.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("UserService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *userBaseService) Get(appuserid string) (utils.Map, error) {
	log.Printf("UserService::FindByCode::  Begin %v", appuserid)

	data, err := p.daoUser.Get(appuserid)
	log.Println("UserService::FindByCode:: End ", err)
	return data, err
}

func (p *userBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("UserService::FindByCode::  Begin ", filter)

	data, err := p.daoUser.Find(filter)
	log.Println("UserService::FindByCode:: End ", data, err)
	return data, err
}

// Create - Create Service
func (p *userBaseService) Create(datauser utils.Map) (utils.Map, error) {

	log.Println("UserService::Insert - Begin")

	// Assign BusinessId
	datauser[business_common.FLD_BUSINESS_ID] = p.businessID

	res, err := p.daoUser.Create(datauser)
	if err != nil {
		log.Println("Business user create Error  ", err)
	}
	log.Println("Business user create  ", res)

	log.Println("UserService::Insert - End ")
	return res, nil
}

// Update - Update Service
func (p *userBaseService) Update(userid string, indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Update - Begin")

	data, err := p.daoUser.Get(userid)
	if err != nil {
		return data, err
	}

	// Delete Key fields
	delete(indata, business_common.FLD_BUSINESS_ID)
	delete(indata, business_common.FLD_USER_ID)

	data, err = p.daoUser.Update(userid, indata)
	log.Println("UserService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *userBaseService) Delete(userid string, deletePermanent bool) error {

	log.Println("UserService::Delete - Begin", userid)

	if deletePermanent {
		result, err := p.daoUser.Delete(userid)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}

		data, err := p.Update(userid, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("UserService::Delete - End")
	return nil
}

func (p *userBaseService) errorReturn(err error) (UserService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
