package tests

import (
	"compass/internal/action"
	"compass/internal/configuration"
	"compass/internal/metricsgroupaction"
	"compass/internal/plugin"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type MetricsGroupActionSuite struct {
	suite.Suite
	DB *gorm.DB

	repository metricsgroupaction.UseCases
	pluginRepo plugin.UseCases
	actionRepo action.UseCases
	mga        metricsgroupaction.MetricsGroupActions
}

func (s *MetricsGroupActionSuite) SetupSuite() {
	os.Setenv("ENV", "TEST")
}

func (s *MetricsGroupActionSuite) BeforeTest(_, _ string) {
	var err error

	s.DB, err = configuration.GetDBConnection("../../migrations")
	require.NoError(s.T(), err)

	s.DB.LogMode(dbLog)

	s.repository = metricsgroupaction.NewMain(s.DB, s.pluginRepo, s.actionRepo)
	clearDatabase(s.DB)
}

func (s *MetricsGroupActionSuite) AfterTest(_, _ string) {
	s.DB.Close()
}

func TestInitMetricsGroupActions(t *testing.T) {
	suite.Run(t, new(MetricsGroupActionSuite))
}

func (s *MetricsGroupActionSuite) TestParseGroupAction() {
	stringReader := strings.NewReader(`{
    "nickname": " ExecutionName ",
    "metricsGroupId": "8800ba87-94e9-443e-9e10-59efe8c58706",
    "actionId": "f1fbe330-c7f6-4215-8311-83015b8df761",
    "executionParameters": {
        "destinationCircleId": "e5b84a9a-340a-49ed-a035-0666506de2d6"
    },
	"configuration": {
		"repeatable": true,
		"numberOfCycles": 0
	}
}`)
	stringReadCloser := ioutil.NopCloser(stringReader)

	res, err := s.repository.ParseGroupAction(stringReadCloser)

	groupID, _ := uuid.Parse("8800ba87-94e9-443e-9e10-59efe8c58706")
	actID, _ := uuid.Parse("f1fbe330-c7f6-4215-8311-83015b8df761")

	require.NoError(s.T(), err)
	require.NotNil(s.T(), res)

	require.Equal(s.T(), "ExecutionName", res.Nickname)
	require.Equal(s.T(), groupID, res.MetricsGroupID)
	require.Equal(s.T(), actID, res.ActionID)
	require.NotNil(s.T(), res.ExecutionParameters)
	require.True(s.T(), res.Configuration.Repeatable)
	require.Equal(s.T(), int16(0), res.Configuration.NumberOfCycles)
}

func (s *MetricsGroupActionSuite) TestParseGroupActionError() {
	stringReader := strings.NewReader(``)
	stringReadCloser := ioutil.NopCloser(stringReader)

	_, err := s.repository.ParseGroupAction(stringReadCloser)

	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSaveMetricsGroupAction() {
	act := newBasicAction()
	group := newBasicMetricGroup()

	s.DB.Create(&group)
	s.DB.Create(&act)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID

	res, err := s.repository.SaveGroupAction(groupAction)

	require.NoError(s.T(), err)
	groupAction.BaseModel = res.BaseModel
	require.Equal(s.T(), groupAction, res)
}

func (s *MetricsGroupActionSuite) TestSaveMetricsGroupActionError() {
	s.DB.Close()
	_, err := s.repository.SaveGroupAction(newBasicGroupAction())

	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestFindByIdMetricsGroupAction() {
	act := newBasicAction()
	group := newBasicMetricGroup()

	s.DB.Create(&group)
	s.DB.Create(&act)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID

	s.DB.Create(&groupAction)

	res, err := s.repository.FindGroupActionById(groupAction.ID.String())
	require.NoError(s.T(), err)
	require.Equal(s.T(), groupAction.ID, res.ID)
}

func (s *MetricsGroupActionSuite) TestFindByIdMetricsGroupActionError() {
	s.DB.Close()
	_, err := s.repository.FindGroupActionById(uuid.New().String())
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestDeleteMetricsGroupAction() {
	act := newBasicAction()
	group := newBasicMetricGroup()

	s.DB.Create(&group)
	s.DB.Create(&act)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID

	s.DB.Create(&groupAction)

	err := s.repository.DeleteGroupAction(groupAction.ID.String())
	require.NoError(s.T(), err)

	var verify metricsgroupaction.MetricsGroupActions
	s.DB.Where("id = ?", groupAction.ID).Find(&verify)

	require.Equal(s.T(), metricsgroupaction.MetricsGroupActions{}, verify)
}

func (s *MetricsGroupActionSuite) TestDeleteMetricsGroupActionError() {
	s.DB.Close()
	err := s.repository.DeleteGroupAction(uuid.New().String())
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestFindAllMetricsGroupAction() {
	act := newBasicAction()
	s.DB.Create(&act)

	group1 := newBasicMetricGroup()
	s.DB.Create(&group1)

	group2 := newBasicMetricGroup()
	s.DB.Create(&group2)

	groupAction1 := newBasicGroupAction()
	groupAction1.ActionID = act.ID
	groupAction1.MetricsGroupID = group1.ID
	groupAction2 := newBasicGroupAction()
	groupAction2.ActionID = act.ID
	groupAction2.MetricsGroupID = group1.ID
	groupAction3 := newBasicGroupAction()
	groupAction3.ActionID = act.ID
	groupAction3.MetricsGroupID = group2.ID
	s.DB.Create(&groupAction1)
	s.DB.Create(&groupAction2)
	s.DB.Create(&groupAction3)

	res, err := s.repository.ListGroupActionExecutionResumeByGroup(group1.ID.String())

	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), res)
	require.Len(s.T(), res, 2)
	require.Equal(s.T(), groupAction1.ID.String(), res[0].Id)
	require.Equal(s.T(), groupAction1.Nickname, res[0].Nickname)
	require.Equal(s.T(), act.Nickname, res[0].ActionType)
	require.Equal(s.T(), "NOT_EXECUTED", res[0].Status)
	require.Nil(s.T(), res[0].StartedAt)
	require.Equal(s.T(), groupAction2.ID.String(), res[1].Id)
	require.Equal(s.T(), groupAction2.Nickname, res[1].Nickname)
	require.Equal(s.T(), act.Nickname, res[1].ActionType)
	require.Equal(s.T(), "NOT_EXECUTED", res[1].Status)
	require.Nil(s.T(), res[1].StartedAt)

}

func (s *MetricsGroupActionSuite) TestFindAllMetricsGroupActionError() {
	s.DB.Close()
	_, err := s.repository.ListGroupActionExecutionResumeByGroup(uuid.New().String())
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestUpdateMetricsGroupAction() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	groupAction.ExecutionParameters = json.RawMessage(`{"update": "eoq"}`)

	res, err := s.repository.UpdateGroupAction(groupAction.ID.String(), groupAction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), groupAction.ExecutionParameters, res.ExecutionParameters)
}

func (s *MetricsGroupActionSuite) TestUpdateMetricsGroupActionError() {
	s.DB.Close()

	_, err := s.repository.UpdateGroupAction("12345", metricsgroupaction.MetricsGroupActions{})
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestCreateNewExecution() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	res, err := s.repository.CreateNewExecution(groupAction.ID.String())
	require.NoError(s.T(), err)

	var executions []metricsgroupaction.ActionsExecutions
	s.DB.Where("group_action_id = ?", groupAction.ID).Find(&executions)

	require.Len(s.T(), executions, 1)
	require.Equal(s.T(), res.ID, executions[0].ID)
	require.Equal(s.T(), "IN_EXECUTION", res.Status)
}

func (s *MetricsGroupActionSuite) TestCreateNewExecutionWrongIDFormat() {
	_, err := s.repository.CreateNewExecution("i'm a wrong format")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestCreateNewExecutionError() {
	s.DB.Close()
	_, err := s.repository.CreateNewExecution(uuid.New().String())
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionFailed() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	s.DB.Create(&execution)

	res, err := s.repository.SetExecutionFailed(execution.ID.String(), "Just Exploded")
	require.NoError(s.T(), err)

	var executions []metricsgroupaction.ActionsExecutions
	s.DB.Where("group_action_id = ?", groupAction.ID).Find(&executions)

	require.Len(s.T(), executions, 1)
	require.Equal(s.T(), res.ID, executions[0].ID)
	require.Equal(s.T(), "FAILED", res.Status)
}

func (s *MetricsGroupActionSuite) TestSetExecutionFailedNotFoundExecution() {
	_, err := s.repository.SetExecutionFailed("i does not exist", "Just Exploded")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionFailedNotInExecution() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	execution.Status = "SUCCESS"
	s.DB.Create(&execution)

	_, err := s.repository.SetExecutionFailed(execution.ID.String(), "Just Exploded")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionFailedError() {
	s.DB.Close()
	_, err := s.repository.SetExecutionFailed(uuid.New().String(), "Ops!")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionSuccess() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	s.DB.Create(&execution)

	res, err := s.repository.SetExecutionSuccess(execution.ID.String(), "Im fine")
	require.NoError(s.T(), err)

	var executions []metricsgroupaction.ActionsExecutions
	s.DB.Where("group_action_id = ?", groupAction.ID).Find(&executions)

	require.Len(s.T(), executions, 1)
	require.Equal(s.T(), res.ID, executions[0].ID)
	require.Equal(s.T(), "SUCCESS", res.Status)
}

func (s *MetricsGroupActionSuite) TestSetExecutionSuccessNotFoundExecution() {
	_, err := s.repository.SetExecutionSuccess("i does not exist", "Im fine")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionSuccessNotInExecution() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	execution.Status = "FAILED"
	s.DB.Create(&execution)

	_, err := s.repository.SetExecutionSuccess(execution.ID.String(), "Im Fine")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestSetExecutionSuccessError() {
	s.DB.Close()
	_, err := s.repository.SetExecutionSuccess(uuid.New().String(), "Ops!")
	require.Error(s.T(), err)
}

func (s *MetricsGroupActionSuite) TestValidateRepeatableActionWithNoExecutionsCanBeExecuted() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	groupAction.Configuration.Repeatable = true
	groupAction.Configuration.NumberOfCycles = 0
	s.DB.Create(&groupAction)

	res := s.repository.ValidateActionCanBeExecuted(groupAction)
	require.True(s.T(), res)
}

func (s *MetricsGroupActionSuite) TestValidateRepeatableActionWithExecutionsCanBeExecuted() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	groupAction.Configuration.Repeatable = true
	groupAction.Configuration.NumberOfCycles = 0
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	execution.Status = "SUCCESS"
	s.DB.Create(&execution)

	res := s.repository.ValidateActionCanBeExecuted(groupAction)
	require.True(s.T(), res)
}

func (s *MetricsGroupActionSuite) TestValidateNotRepeatableActionWithNoExecutionsCanBeExecuted() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	groupAction.Configuration.Repeatable = false
	groupAction.Configuration.NumberOfCycles = 1
	s.DB.Create(&groupAction)

	res := s.repository.ValidateActionCanBeExecuted(groupAction)
	require.True(s.T(), res)
}

func (s *MetricsGroupActionSuite) TestValidateNotRepeatableActionWithExecutionsCanBeExecuted() {
	act := newBasicAction()
	s.DB.Create(&act)

	group := newBasicMetricGroup()
	s.DB.Create(&group)

	groupAction := newBasicGroupAction()
	groupAction.ActionID = act.ID
	groupAction.MetricsGroupID = group.ID
	groupAction.Configuration.Repeatable = false
	groupAction.Configuration.NumberOfCycles = 1
	s.DB.Create(&groupAction)

	execution := newBasicActionExecution()
	execution.GroupActionId = groupAction.ID
	execution.Status = "SUCCESS"
	s.DB.Create(&execution)

	res := s.repository.ValidateActionCanBeExecuted(groupAction)
	require.False(s.T(), res)
}

func (s *MetricsGroupActionSuite) TestValidateActionCanBeExecutedError() {
	s.DB.Close()
	res := s.repository.ValidateActionCanBeExecuted(newBasicGroupAction())
	require.False(s.T(), res)
}
