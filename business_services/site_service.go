package business_services

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business-repository/business_common"
	"github.com/zapscloud/golib-business-repository/business_repository"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// SiteService - Sites Service structure
type SiteService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	GetDetails(siteid string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(siteid string, indata utils.Map) (utils.Map, error)
	Delete(siteid string) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// SiteBaseService - Sites Service structure
type siteBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoSite     business_repository.SiteDao
	daoBusiness platform_repository.BusinessDao
	child       SiteService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewSiteService(props utils.Map) (SiteService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("SiteService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := siteBaseService{}
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

// SiteBaseService - Close all the services
func (p *siteBaseService) EndService() {
	log.Printf("EndSiteService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *siteBaseService) initializeService() {
	log.Printf("SiteService:: GetBusinessDao ")
	p.daoSite = business_repository.NewSiteDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *siteBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("SiteService::FindAll - Begin")

	daoSite := p.daoSite
	response, err := daoSite.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("SiteService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *siteBaseService) GetDetails(site_id string) (utils.Map, error) {
	log.Printf("SiteService::FindByCode::  Begin %v", site_id)

	data, err := p.daoSite.Get(site_id)
	log.Println("SiteService::FindByCode:: End ", err)
	return data, err
}

func (p *siteBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("SiteService::FindByCode::  Begin ", filter)

	data, err := p.daoSite.Find(filter)
	log.Println("SiteService::FindByCode:: End ", data, err)
	return data, err
}

func (p *siteBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[business_common.FLD_APP_SITE_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("site")
		log.Println("Unique Site ID", uid)
		indata[business_common.FLD_APP_SITE_ID] = uid
		dataval = indata[business_common.FLD_APP_SITE_ID]
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Site ID:", dataval)

	_, err := p.daoSite.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Site ID !", ErrorDetail: "Given Site ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoSite.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *siteBaseService) Update(site_id string, indata utils.Map) (utils.Map, error) {

	log.Println("SiteService::Update - Begin")

	data, err := p.daoSite.Get(site_id)
	if err != nil {
		return data, err
	}

	data, err = p.daoSite.Update(site_id, indata)
	log.Println("SiteService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *siteBaseService) Delete(site_id string) error {

	log.Println("SiteService::Delete - Begin", site_id)

	daoSite := p.daoSite
	result, err := daoSite.Delete(site_id)
	if err != nil {
		return err
	}

	log.Printf("SiteService::Delete - End %v", result)
	return nil
}

func (p *siteBaseService) errorReturn(err error) (SiteService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
