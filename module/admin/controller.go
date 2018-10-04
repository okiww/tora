package admin

import (
	"errors"
	"fmt"
	"net/http"
	"okkybudiman/data"
	dataModel "okkybudiman/data/model"
	"strconv"

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

func (ctrl *Controller) CreateTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var req testRequest
	var test dataModel.Test
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
	//save data
	if err := db.Where("name = ?", req.Name).Find(&test).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "test name already exist",
		})
		return
	}
	test = dataModel.Test{
		Name:          req.Name,
		Description:   req.Description,
		TotalQuestion: req.TotalQuestion,
	}

	db.Save(&test)

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "success create test",
	})
	return
}

func (ctrl *Controller) CreateQuestion(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var req questionRequest
	var question dataModel.Question
	var choices dataModel.QuestionChoice
	var test dataModel.Test
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
	}

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

	uid, err := uuid.FromString(req.TestID)
	var count int
	//save data
	if err := db.Where("id =?", uid).Find(&test).Error; err == nil {
		db.Model(&question).Where("test_id =?", uid).Count(&count)

		//handle if question is full
		if count == test.TotalQuestion {
			c.JSON(http.StatusOK, gin.H{
				"status":  http.StatusOK,
				"message": "question is full for this test",
			})
			return
		}
		totalLeft := test.TotalQuestion - count
		for k, q := range req.Questions {
			question = dataModel.Question{
				Question: q.Question,
				Answer:   q.Answer,
				TestID:   uid,
			}

			//handle total question
			if k < totalLeft {
				db.Save(&question)
				for key, v := range q.Choices {
					choices = dataModel.QuestionChoice{
						Choice:     v.Choice,
						Key:        strconv.Itoa(key + 1),
						QuestionID: question.ID,
					}

					db.Save(&choices)
				}
			}
		}
		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "success create question",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Test",
	})
	return
}

func (ctrl *Controller) GetTestByID(c *gin.Context) {
}

func (ctrl *Controller) GetListTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var tests []dataModel.Test
	var responses []testResponse

	if err := db.Find(&tests).Error; err == nil {

		for _, v := range tests {
			res := testResponse{
				ID:            v.ID,
				Name:          v.Name,
				Description:   v.Description,
				TotalQuestion: v.TotalQuestion,
			}
			responses = append(responses, res)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success get list test",
			"data":    responses,
			"total":   len(tests),
		})

		return
	}
}

func (ctrl *Controller) GetQuestion(c *gin.Context) {
}

func (ctrl *Controller) GetParticipant(c *gin.Context) {
}

func (ctrl *Controller) UpdateTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var test dataModel.Test
	var req updateTestRequest

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

	uid, err := uuid.FromString(req.TestID)
	if err := db.Where("id = ?", uid).First(&test).Error; err == nil {
		test.Name = req.Name
		test.Description = req.Description
		test.TotalQuestion = req.TotalQuestion

		db.Save(&test)

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "success update test",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Test",
	})
	return
}

func (ctrl *Controller) UpdateQuestion(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var question dataModel.Question
	var req updateQuestionRequest

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

	uid, err := uuid.FromString(req.QuestionID)
	if err := db.Where("id = ?", uid).First(&question).Error; err == nil {
		question.Question = req.Question
		question.Answer = req.Answer

		db.Save(&question)

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "success update question",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Question",
	})
	return
}

func (ctrl *Controller) UpdateChoice(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var questionChoice dataModel.QuestionChoice
	var req updateQuestionChoiceRequest

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

	uid, err := uuid.FromString(req.ChoiceID)
	if err := db.Where("id = ?", uid).First(&questionChoice).Error; err == nil {
		questionChoice.Choice = req.Choice

		db.Save(&questionChoice)

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "success update choice",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Choice",
	})
	return
}

func (ctrl *Controller) DeleteTest(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var test dataModel.Test
	var questions []dataModel.Question
	var questionChoice dataModel.QuestionChoice
	var req deleteTestRequest

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

	uid, err := uuid.FromString(req.TestID)
	if err := db.Where("id =?", uid).Find(&test).Error; err == nil {
		db.Delete(&test)

		if err := db.Where("test_id =?", uid).Find(&questions).Error; err == nil {
			for _, q := range questions {
				db.Delete(&questionChoice).Where("question_id =?", q.ID)
			}

			db.Delete(&questions).Where("test_id =?", uid)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success delete test",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Test",
	})
	return
}

func (ctrl *Controller) DeleteQuestion(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var question dataModel.Question
	var questionChoice dataModel.QuestionChoice
	var req deleteQuestionRequest

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

	uid, err := uuid.FromString(req.QuestionID)

	if err := db.Where("id =?", uid).Find(&question).Error; err == nil {
		db.Delete(&question).Where("id =?", uid)
		db.Where("question_id =?", uid).Delete(&questionChoice)

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success delete question",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Question",
	})
	return
}

func (ctrl *Controller) DeleteChoice(c *gin.Context) {
	db, err := ctrl.dbFactory.DBConnection()
	if err != nil {
		fmt.Println("err")
		glog.Errorf("Failed to open db connection: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var questionChoices []dataModel.QuestionChoice
	var req deleteChoiceRequest

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

	uid, err := uuid.FromString(req.QuestionID)
	uid2, err := uuid.FromString(req.ChoiceID)

	if err := db.Where("question_id =?", uid).Find(&questionChoices).Error; err == nil {

		totalChoice := len(questionChoices)

		if totalChoice > 2 {
			db.Where("id =?", uid2).Delete(&questionChoices)

			c.JSON(http.StatusOK, gin.H{
				"status":  http.StatusOK,
				"message": "success delete choice",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "minimum total choices is 2. add choice first and then delete one",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusNotFound,
		"message": "cannot find Choice",
	})
	return
}
