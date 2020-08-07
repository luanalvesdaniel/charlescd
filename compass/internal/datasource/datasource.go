package datasource

import (
	"compass/internal/util"
	"encoding/json"
	"errors"
	"log"
	"plugin"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type DataSource struct {
	util.BaseModel
	Name        string      `json:"name"`
	PluginID    uuid.UUID   `json:"pluginId"`
	Health      bool        `json:"health"`
	Data        interface{} `json:"data"`
	WorkspaceID string      `json:"workspaceId"`
	Deleted     bool
	DeletedAt   time.Time
}

type MetricList struct {
	Data []string
}

func (main Main) FindAllByWorkspace(workspaceID string) ([]DataSource, error) {
	dataSources := []DataSource{}
	db := main.db.Where("workspace_id = ? AND deleted = false", workspaceID).Find(&dataSources)
	if db.Error != nil {
		return []DataSource{}, db.Error
	}
	return dataSources, nil
}

func (main Main) findById(id string) (DataSource, error) {
	dataSource := DataSource{}
	result := main.db.Where("id = ?", id).First(&dataSource)
	if result.Error != nil && gorm.IsRecordNotFoundError(result.Error) {
		return DataSource{}, errors.New("Not found")
	}

	if result.Error != nil {
		return DataSource{}, result.Error
	}
	return dataSource, nil
}

func (main Main) verifyHealthAtWorkspace(workspaceId string) (bool, error) {
	var count int8
	result := main.db.Where("workspace_id = ? AND health", workspaceId).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count != 0, nil
}

func (main Main) Delete(id string, workspaceID string) error {
	if _, err := main.findById(id); err != nil {
		return err
	}

	db := main.db.Model(&DataSource{}).Where("id = ?", id).Update(DataSource{Deleted: true, DeletedAt: time.Now()})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func (main Main) GetMetrics(dataSourceID, name string) (MetricList, error) {
	dataSourceResult, err := main.findById(dataSourceID)
	if err != nil {
		return MetricList{}, err
	}

	pluginResult, err := main.pluginMain.FindById(dataSourceResult.PluginID.String())
	if err != nil {
		return MetricList{}, err
	}

	plugin, err := plugin.Open(pluginResult.Src)
	if err != nil {
		return MetricList{}, err
	}

	getList, err := plugin.Lookup("GetLists")
	if err != nil {
		return MetricList{}, err
	}

	configurationData, _ := json.Marshal(dataSourceResult.Data)
	list, err := getList.(func(configurationData []byte) (MetricList, error))(configurationData)
	if err != nil {
		return MetricList{}, err
	}

	return list, nil
}
func (main Main) SetAsHealth(id string, workspaceID string) error {
	if hasHealth, err := main.verifyHealthAtWorkspace(workspaceID); err != nil || hasHealth {
		log.Print(err)
		return errors.New("Cannot set as Health")
	}

	db := main.db.Model(&DataSource{}).Where("id = ?", id).Update(DataSource{Health: true})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

/* func (main Main) Save(dataSource DataSource) error {
	db := main.db.Model(DataSource{}).Where("id = ?", id).Update("deleted", true)
	if gorm.IsRecordNotFoundError(db.Error) {
		return errors.New("Not Found")
	}
	if db.Error != nil {
		return db.Error
	}
	return nil
} */