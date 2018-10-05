package user

import (
	"errors"
	"fmt"
	"net/http"
	"okkybudiman/data"
	dataModel "okkybudiman/data/model"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	uuid "github.com/satori/go.uuid"
	validator "gopkg.in/go-playground/validator.v8"
)

type Controller struct {
	dbFactory *data.DBFactory
}

func NewController(dbFactory *data.DBFactory) (*Controller, error) {
	if dbFactory == nil {
		return nil, errors.New("failed to instantiate rate controller")
	}

	return &Controller{dbFactory: dbFactory}, nil
}

func (ctrl *Controller) AttempTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	claims := jwt.ExtractClaims(c)
	name := claims["id"].(string)
	var user dataModel.User
	db.Where("name = ?", name).Find(&user)
	userId := user.ID

	var req attempRequest
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		var errors []string
		ve, ok := err.(validator.ValidationErrors)
		if ok {
			for _, v := range ve {
				errors = append(errors, fmt.Sprintf("%s is %s", v.Field, v.Tag))
			}
		} else {
			errors = append(errors, err.Error())
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	testID, _ := uuid.FromString(req.TestID)

	//set timezone,
	now := time.Now()
	attemptTest := dataModel.UserAttemptTest{
		UserID:     userId,
		TestID:     testID,
		IsFinished: false,
		StartTest:  now,
		EndTest:    now,
	}

	db.Save(&attemptTest)

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "success attempt test",
	})
	return
}

func (ctrl *Controller) AnswerTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	claims := jwt.ExtractClaims(c)
	name := claims["id"].(string)
	var user dataModel.User
	db.Where("name = ?", name).Find(&user)
	userId := user.ID

	var test dataModel.Test
	var userAttempt dataModel.UserAttemptTest
	var req answerRequest
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		var errors []string
		ve, ok := err.(validator.ValidationErrors)
		if ok {
			for _, v := range ve {
				errors = append(errors, fmt.Sprintf("%s is %s", v.Field, v.Tag))
			}
		} else {
			errors = append(errors, err.Error())
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	totalRightAnswer := 0
	totalWrongAnswer := 0
	totalNotAnswer := 0
	calculatePoint := 0
	//save data
	testID, _ := uuid.FromString(req.TestID)
	if err := db.Where("id = ?", testID).Find(&test).Error; err == nil {
		for _, v := range req.Answers {
			var question dataModel.Question

			questionID, _ := uuid.FromString(v.QuestionID)
			var point int
			db.Where("id = ?", questionID).Find(&question)
			if question.Answer == v.Answer {
				point = 4
				totalRightAnswer = totalRightAnswer + 1
			} else {
				point = -2
				totalWrongAnswer = totalWrongAnswer + 1
			}

			if v.Answer == "" {
				point = 0
				totalNotAnswer = totalNotAnswer + 1
			}

			calculatePoint = calculatePoint + point

			answer := dataModel.UserAnswer{
				UserID:     userId,
				TestID:     testID,
				QuestionID: questionID,
				Answer:     v.Answer,
				Point:      point,
			}

			db.Save(&answer)
		}
	}
	//update score
	score := dataModel.UserScore{
		UserID:             userId,
		TestID:             testID,
		TotalRightAnswered: totalRightAnswer,
		TotalWrongAnswered: totalWrongAnswer,
		TotalNotAnswered:   totalNotAnswer,
		Score:              calculatePoint,
	}

	db.Save(&score)
	//update tb user_attempt_test
	db.Where("test_id = ? AND user_id = ?", testID, userId).Find(&userAttempt)
	t1 := userAttempt.StartTest
	t2 := time.Now()

	diff := t2.Sub(t1)
	out := time.Time{}.Add(diff)

	userAttempt.IsFinished = true
	userAttempt.EndTest = t2
	userAttempt.FinishTime = out.Format("15:04:05")

	db.Save(&userAttempt)

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "success save data",
	})
	return
}

func (ctrl *Controller) Result(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	claims := jwt.ExtractClaims(c)
	name := claims["id"].(string)
	var user dataModel.User
	db.Where("name = ?", name).Find(&user)
	userId := user.ID

	var userScore dataModel.UserScore
	var userAttempt dataModel.UserAttemptTest
	id := c.Param("id")
	testID, _ := uuid.FromString(id)
	if err := db.Where("test_id = ? AND user_id = ?", testID, userId).Find(&userScore).Error; err == nil {
		db.Where("test_id = ? AND user_id = ?", testID, userId).Find(&userAttempt)
		data := result{
			ID:                 userScore.ID,
			UserID:             userScore.UserID,
			Name:               name,
			TotalRightAnswered: userScore.TotalRightAnswered,
			TotalWrongAnswered: userScore.TotalWrongAnswered,
			TotalNotAnswered:   userScore.TotalNotAnswered,
			Score:              userScore.Score,
			TimeComplete:       userAttempt.FinishTime,
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success get data",
			"results": data,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "cannot find data",
		"result":  nil,
	})
}
