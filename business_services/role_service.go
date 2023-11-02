package business_services

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

// RoleService - Roles Service structure
type RoleService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(roleid string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(roleid string, indata utils.Map) (utils.Map, error)
	Delete(roleid string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// RoleBaseService - Roles Service structure
type roleBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoRole     business_repository.RoleDao
	daoBusiness platform_repository.BusinessDao
	child       RoleService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}
func NewRoleService(props utils.Map) (RoleService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("RoleService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := roleBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		log.Fatal(err)
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

// RoleBaseService - Close all the services
func (p *roleBaseService) EndService() {
	log.Printf("EndRoleService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *roleBaseService) initializeService() {
	log.Printf("RoleMongoService:: GetBusinessDao ")
	p.daoRole = business_repository.NewRoleDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *roleBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("RoleService::FindAll - Begin")

	daoRole := p.daoRole
	response, err := daoRole.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("RoleService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *roleBaseService) Get(approleid string) (utils.Map, error) {
	log.Printf("RoleService::FindByCode::  Begin %v", approleid)

	data, err := p.daoRole.GetDetails(approleid)
	log.Println("RoleService::FindByCode:: End ", err)
	return data, err
}

func (p *roleBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("RoleService::FindByCode::  Begin ", filter)

	data, err := p.daoRole.Find(filter)
	log.Println("RoleService::FindByCode:: End ", data, err)
	return data, err
}

func (p *roleBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	var roleId string

	dataval, dataok := indata[business_common.FLD_ROLE_ID]
	if dataok {
		roleId = dataval.(string)
	} else {
		roleId = utils.GenerateUniqueId("role")
		log.Println("Unique Role ID", roleId)
		indata[business_common.FLD_ROLE_ID] = roleId
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Role ID:", roleId)

	_, err := p.daoRole.GetDetails(roleId)
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Role ID !", ErrorDetail: "Given Role ID already exist"}
		return nil, err
	}

	insertResult, err := p.daoRole.Create(indata)
	if err != nil {
		return nil, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *roleBaseService) Update(roleid string, indata utils.Map) (utils.Map, error) {

	log.Println("RoleService::Update - Begin")

	data, err := p.daoRole.GetDetails(roleid)
	if err != nil {
		return data, err
	}

	// Remove the fields which should not be updated in database
	delete(indata, db_common.FLD_IS_AUTO_GENERATED)

	data, err = p.daoRole.Update(roleid, indata)
	log.Println("RoleService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *roleBaseService) Delete(roleid string, delete_permanent bool) error {

	log.Println("RoleService::Delete - Begin", roleid, delete_permanent)

	daoRole := p.daoRole
	if delete_permanent {
		result, err := daoRole.Delete(roleid)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(roleid, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("RoleService::Delete - End")
	return nil
}

func (p *roleBaseService) errorReturn(err error) (RoleService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
