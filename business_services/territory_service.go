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

// TerritoryService - Territorys Service structure
type TerritoryService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	GetDetails(territoryid string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(territoryid string, indata utils.Map) (utils.Map, error)
	Delete(territoryid string) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// TerritoryBaseService - Territorys Service structure
type territoryBaseService struct {
	db_utils.DatabaseService
	dbRegion     db_utils.DatabaseService
	daoTerritory business_repository.TerritoryDao
	daoBusiness  platform_repository.BusinessDao
	child        TerritoryService
	businessID   string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewTerritoryService(props utils.Map) (TerritoryService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Print("TerritoryService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := territoryBaseService{}
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

// TerritoryBaseService - Close all the services
func (p *territoryBaseService) EndService() {
	log.Printf("EndTerritoryService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *territoryBaseService) initializeService() {
	log.Printf("TerritoryService:: GetBusinessDao ")
	p.daoTerritory = business_repository.NewSiteDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *territoryBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("TerritoryService::FindAll - Begin")

	daoTerritory := p.daoTerritory
	response, err := daoTerritory.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("TerritoryService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *territoryBaseService) GetDetails(territory_id string) (utils.Map, error) {
	log.Printf("TerritoryService::FindByCode::  Begin %v", territory_id)

	data, err := p.daoTerritory.Get(territory_id)
	log.Println("TerritoryService::FindByCode:: End ", err)
	return data, err
}

func (p *territoryBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("TerritoryService::FindByCode::  Begin ", filter)

	data, err := p.daoTerritory.Find(filter)
	log.Println("TerritoryService::FindByCode:: End ", data, err)
	return data, err
}

func (p *territoryBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[business_common.FLD_APP_TERRITORY_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("territory")
		log.Println("Unique Territory ID", uid)
		indata[business_common.FLD_APP_TERRITORY_ID] = uid
		dataval = indata[business_common.FLD_APP_TERRITORY_ID]
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Territory ID:", dataval)

	_, err := p.daoTerritory.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Territory ID !", ErrorDetail: "Given Territory ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoTerritory.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *territoryBaseService) Update(territory_id string, indata utils.Map) (utils.Map, error) {

	log.Println("TerritoryService::Update - Begin")

	data, err := p.daoTerritory.Get(territory_id)
	if err != nil {
		return data, err
	}

	data, err = p.daoTerritory.Update(territory_id, indata)
	log.Println("TerritoryService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *territoryBaseService) Delete(territory_id string) error {

	log.Println("TerritoryService::Delete - Begin", territory_id)

	daoTerritory := p.daoTerritory
	result, err := daoTerritory.Delete(territory_id)
	if err != nil {
		return err
	}

	log.Printf("TerritoryService::Delete - End %v", result)
	return nil
}

func (p *territoryBaseService) errorReturn(err error) (TerritoryService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
