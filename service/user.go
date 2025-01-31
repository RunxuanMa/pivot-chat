package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Pivot-Studio/pivot-chat/util"
	"gorm.io/gorm"

	"github.com/Pivot-Studio/pivot-chat/conf"
	"github.com/Pivot-Studio/pivot-chat/constant"
	"github.com/Pivot-Studio/pivot-chat/dao"
	"github.com/Pivot-Studio/pivot-chat/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

func Auth(Email string, Password string) bool {
	user := &model.User{}
	err := dao.RS.GetUserByEmail(user, Email)
	if err != nil {
		logrus.Fatalf("[Service.Auth] GetUserByEmail file %+v", err)
		return false
	}
	return util.ComparePassword(user.Password, Password)
}
func Register(ctx *gin.Context, user *model.User, captcha string) (err error) {
	//邮箱验证码部分
	res, err := dao.Cache.Get(context.Background(), user.Email).Result()
	if err != nil {
		return err
	}
	if res != captcha {
		return constant.CaptchaErr
	}
	err = dao.RS.GetUserByEmail(&model.User{Email: user.Email}, user.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else if err == nil {
		return errors.New("该邮箱已被注册")
	}
	err = dao.RS.CreateUser([]*model.User{user})
	if err != nil {
		return err
	}
	return nil
}

// 生成验证码
func CreatCode() (code string) {
	rand.Seed(time.Now().Unix())
	code = fmt.Sprintf("%6v", rand.Intn(600000))
	return
}

// 发送验证码
func SendEmail(ctx context.Context, email string, captcha string) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", conf.C.EmailServer.Email)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "邮箱验证")
	content := strings.Replace(emailContent, "VerifyCodePlace", captcha, -1)
	m.SetBody("text/html", content)
	err = d.DialAndSend(m)
	if err != nil {
		logrus.Error("[SendEmail] send to email:%s err:%+v", email, err)
		return err
	}
	return nil
}

// 将验证码存入redis
func CaptchaLogic(ctx *gin.Context, code, email string) {
	codeKey := email
	dao.Cache.Set(ctx, codeKey, code, time.Minute*5) //存入redis 有效5min
}

// 比较验证码
func CaptchaCheck(ctx *gin.Context, input string, email string) bool {
	code := dao.Cache.Get(ctx, email).String() //对比验证码是否一致
	return code == input
}

func ChgPwd(ctx *gin.Context, email string, oldPwd string, newPwd string) error {
	return dao.RS.ChangeUserPwd(email, oldPwd, newPwd)
}

func FindUserById(ctx *gin.Context, userid int64) (err error, data map[string]interface{}) {
	user := new(model.User)
	user.UserId = userid
	err = dao.RS.GetUserbyId(user)
	if err != nil {
		logrus.Fatalf("[Service.FindUserById] FindUserById %+v", err)
		return err, nil
	}
	data["username"] = user.UserName
	data["user_id"] = user.UserId
	data["email"] = user.Email
	return nil, data
}
