package business_services

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business/business_common"
	"github.com/zapscloud/golib-business/business_repository"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// AccessService - Accesss Service structure
type AccessService interface {
	List(sys_filter string, filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(access_id string) (utils.Map, error)

	GrantPermission(indata utils.Map) (utils.Map, error)
	RevokePermission(access_id string) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// AccessBaseService - Accesss Service structure
type accessBaseService struct {
	db_utils.DatabaseService
	dbRegion  db_utils.DatabaseService
	daoAccess business_repository.AccessDao
	daoUser   business_repository.UserDao
	daoRole   business_repository.RoleDao
	daoSite   business_repository.SiteDao

	daoSysUser  platform_repository.SysUserDao
	daoSysRole  platform_repository.SysRoleDao
	daoBusiness platform_repository.BusinessDao
	child       AccessService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewAccessService(props utils.Map) (AccessService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"
	log.Printf("AccessService :: Start")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := accessBaseService{}
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

	// Initialize other Service
	p.initializeService()

	// Verify the given businessId is exist
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

func (p *accessBaseService) EndService() {
	log.Printf("EndAccessService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *accessBaseService) initializeService() {
	log.Printf("AccessService:: GetBusinessDao ")
	p.daoAccess = business_repository.NewAccessDao(p.dbRegion.GetClient(), p.businessID)
	p.daoUser = business_repository.NewUserDao(p.dbRegion.GetClient(), p.businessID)
	p.daoRole = business_repository.NewRoleDao(p.dbRegion.GetClient(), p.businessID)
	p.daoSite = business_repository.NewSiteDao(p.dbRegion.GetClient(), p.businessID)

	p.daoSysUser = platform_repository.NewSysUserDao(p.GetClient())
	p.daoSysRole = platform_repository.NewSysRoleDao(p.GetClient())
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

func (p *accessBaseService) getServiceModuleCode() string {
	return business_common.GetServiceModuleCode() + "05"
}

// List - List All records
func (p *accessBaseService) List(sys_filter string, filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("AccessService::FindAll - Begin")

	daoAccess := p.daoAccess
	response, err := daoAccess.List(sys_filter, filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("AccessService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *accessBaseService) Get(access_id string) (utils.Map, error) {
	log.Printf("AccessService::FindByCode::  Begin %v", access_id)

	data, err := p.daoAccess.Get(access_id)
	log.Println("AccessService::FindByCode:: End ", err)
	return data, err
}

func (p *accessBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("AccessService::FindByCode::  Begin ", filter)

	data, err := p.daoAccess.Find(filter)
	log.Println("AccessService::FindByCode:: End ", data, err)
	return data, err
}

// Update - Update Service
func (p *accessBaseService) GrantPermission(indata utils.Map) (utils.Map, error) {

	funcode := p.getServiceModuleCode() + "01"

	log.Println("AccessService::Update - Begin")

	access_key := ""
	if valUserId, okUserId := indata[business_common.FLD_USER_ID]; !okUserId {
		log.Println("GrantPermission: UserId not found  ", valUserId)
		err := &utils.AppError{ErrorCode: funcode + "01", ErrorMsg: "UserId not found ", ErrorDetail: "UserId not found "}
		return indata, err
	} else if _, err := p.daoUser.Get(valUserId.(string)); err != nil {
		log.Println("GrantPermission: UserId not found  ", valUserId)
		err := &utils.AppError{ErrorCode: funcode + "02", ErrorMsg: "UserId not found ", ErrorDetail: "UserId not found "}
		return indata, err
	} else {
		access_key += valUserId.(string)
	}

	valRoleId, okRoleId := indata[business_common.FLD_ROLE_ID]
	valSysRoleId, okSysRoleId := indata["app_role_id"]

	if !okRoleId && !okSysRoleId {
		log.Println("GrantPermission: Missing RoleId ")
		err := &utils.AppError{ErrorCode: funcode + "03", ErrorMsg: "Missing RoleId ", ErrorDetail: "Missing RoleId "}
		return indata, err
	}

	if okRoleId {
		dataRole, err := p.daoRole.GetDetails(valRoleId.(string))
		log.Println("GrantPermission: RoleId not found  ", dataRole, err)
		if err != nil {
			err := &utils.AppError{ErrorCode: funcode + "04", ErrorMsg: "RoleId not found ", ErrorDetail: "RoleId not found "}
			return indata, err
		}
	}

	if okSysRoleId {
		dataRole, err := p.daoSysRole.Get(valSysRoleId.(string))
		log.Println("GrantPermission: RoleId not found  ", dataRole, err)
		if err != nil {
			err := &utils.AppError{ErrorCode: funcode + "04", ErrorMsg: "RoleId not found ", ErrorDetail: "App RoleId not found "}
			return indata, err
		}
	}

	if valSiteId, okSiteId := indata[business_common.FLD_APP_SITE_ID]; !okSiteId {
		// Ignore Site Id Field
		access_key = "-" + access_key
	} else if _, err := p.daoSite.Get(valSiteId.(string)); err != nil {
		log.Println("GrantPermission: RoleId not found  ", valSiteId)
		err := &utils.AppError{ErrorCode: funcode + "04", ErrorMsg: "UserId not found ", ErrorDetail: "UserId not found "}
		return indata, err
	} else {
		access_key += valSiteId.(string)
	}

	access_id := utils.GenerateChecksumId("aces", access_key)
	dataAccess, err := p.daoAccess.Get(access_id)
	if err != nil {
		return dataAccess, err
	}

	indata[business_common.FLD_APP_ACCESS_ID] = access_id

	dataAccess, err = p.daoAccess.GrantPermission(indata)
	log.Println("AccessService::Update - End ")
	return dataAccess, err
}

// RevokePermission - RevokePermission Service
func (p *accessBaseService) RevokePermission(access_id string) error {

	log.Println("AccessService::RevokePermission - Begin", access_id)

	daoUser := p.daoAccess
	result, err := daoUser.RevokePermission(access_id)
	if err != nil {
		return err
	}

	log.Printf("UserService::Delete - End %v", result)
	return nil
}

func (p *accessBaseService) errorReturn(err error) (AccessService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
