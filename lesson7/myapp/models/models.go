package models

import (
	//"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	_DB_NAME      = "root:123456@/test?charset=utf8"
	_MYSQL_DRIVER = "mysql"
)

// 分类
type Category struct {
	Id              int64
	Title           string
	Created         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	TopicTime       time.Time `orm:"index"`
	TopicCount      int64
	TopicLastUserId int64
}

// 文章
type Topic struct {
	Id              int64
	Uid             int64
	Title           string
	Category        string
	Lables          string
	Content         string `orm:"size(5000)"`
	Attachment      string
	Created         time.Time `orm:"index"`
	Updated         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	Author          string
	ReplyTime       time.Time `orm:"index"`
	ReplyCount      int64
	ReplyLastUserId int64
}

// 评论
type Comment struct {
	Id      int64
	Tid     int64
	Name    string
	Content string    `orm:"size(1000)"`
	Created time.Time `orm:"index"`
}

func RegisterDB() {
	// 注册模型
	orm.RegisterModel(new(Category), new(Topic), new(Comment))
	orm.RegisterDriver(_MYSQL_DRIVER, orm.DRMySQL)

	// 注册默认数据库
	orm.RegisterDataBase("default", _MYSQL_DRIVER, _DB_NAME, 10)
}

func AddCategory(name string) error {
	o := orm.NewOrm()

	cate := &Category{
		Title:     name,
		Created:   time.Now(),
		TopicTime: time.Now(),
	}

	// 查询数据
	qs := o.QueryTable("category")
	err := qs.Filter("title", name).One(cate)
	if err == nil {
		return err
	}

	// 插入数据
	_, err = o.Insert(cate)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCategory(id string) error {
	cid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	cate := &Category{Id: cid}
	_, err = o.Delete(cate)
	return err
}

func GetAllCategories() ([]*Category, error) {
	o := orm.NewOrm()

	cates := make([]*Category, 0)

	qs := o.QueryTable("category")
	_, err := qs.All(&cates)
	return cates, err
}

func AddTopic(title, category, lable, content, attachment string) error {
	// 处理标签
	lable = "$" + strings.Join(strings.Split(lable, " "), "#$") + "#"

	o := orm.NewOrm()

	topic := &Topic{
		Title:      title,
		Category:   category,
		Lables:     lable,
		Content:    content,
		Attachment: attachment,
		Created:    time.Now(),
		Updated:    time.Now(),
		ReplyTime:  time.Now(),
	}
	_, err := o.Insert(topic)
	if err != nil {
		return err
	}

	// 更新分类统计
	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("title", category).One(cate)
	if err == nil {
		// 如果不存在我们就直接忽略，只当分类存在时进行更新
		cate.TopicCount++
		_, err = o.Update(cate)
	}

	return err

}

func GetTopic(tid string) (*Topic, error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()

	topic := new(Topic)

	qs := o.QueryTable("topic")
	err = qs.Filter("id", tidNum).One(topic)
	if err != nil {
		return nil, err
	}

	topic.Views++
	_, err = o.Update(topic)
	if err != nil {
		return nil, err
	}

	topic.Lables = strings.Replace(strings.Replace(
		topic.Lables, "#", " ", -1), "$", "", -1)

	return topic, nil
}

func ModifyTopic(tid, title, category, lable, content, attachment string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}

	// 处理标签
	lable = "$" + strings.Join(strings.Split(lable, " "), "#$") + "#"

	o := orm.NewOrm()
	topic := &Topic{Id: tidNum}
	var oldCate, oldAttach string
	if o.Read(topic) == nil {
		oldCate = topic.Category
		oldAttach = topic.Attachment
		topic.Title = title
		topic.Category = category
		topic.Lables = lable
		topic.Content = content
		topic.Attachment = attachment
		topic.Updated = time.Now()
		_, err = o.Update(topic)
		if err != nil {
			return err
		}
	}

	// 删除旧的附件
	if len(oldAttach) > 0 && oldAttach != topic.Attachment {
		os.Remove(path.Join("attachment", oldAttach))
	}

	if len(oldCate) > 0 {
		cate := new(Category)
		qs := o.QueryTable("category")
		err = qs.Filter("title", oldCate).One(cate)
		if err == nil {
			cate.TopicCount--
			_, err = o.Update(cate)
		}
	}

	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("title", category).One(cate)
	if err == nil {
		cate.TopicCount++
		_, err = o.Update(cate)
	}

	return nil
}

func DeleteTopic(tid string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	topic := &Topic{Id: tidNum}
	_, err = o.Delete(topic)
	return err
}

func GetAllTopics(category string, lable string, isDesc bool) (topics []*Topic, err error) {
	o := orm.NewOrm()

	topics = make([]*Topic, 0)

	qs := o.QueryTable("topic")
	if isDesc {
		if len(category) > 0 {
			qs = qs.Filter("category", category)
		}
		if len(lable) > 0 {
			qs = qs.Filter("lables__contains", "$"+lable+"#")
		}
		_, err = qs.OrderBy("-created").All(&topics)
	} else {
		_, err = qs.All(&topics)
	}
	return topics, err
}

func AddReply(tid, nickname, content string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}

	reply := &Comment{
		Tid:     tidNum,
		Name:    nickname,
		Content: content,
		Created: time.Now(),
	}
	o := orm.NewOrm()
	_, err = o.Insert(reply)
	if err != nil {
		return nil
	}

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		topic.ReplyTime = time.Now()
		topic.ReplyCount++
		_, err = o.Update(topic)
	}

	return err
}

func GetAllReplies(tid string) (replies []*Comment, err error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	replies = make([]*Comment, 0)

	o := orm.NewOrm()
	qs := o.QueryTable("comment")
	_, err = qs.Filter("tid", tidNum).All(&replies)
	return replies, err
}

func DeleteReply(rid string) error {
	ridNum, err := strconv.ParseInt(rid, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	var tidNum int64
	reply := &Comment{Id: ridNum}
	if o.Read(reply) == nil {
		tidNum = reply.Tid
		_, err = o.Delete(reply)
		if err != nil {
			return err
		}
	}

	replies := make([]*Comment, 0)
	qs := o.QueryTable("comment")
	_, err = qs.Filter("tid", tidNum).OrderBy("-created").All(&replies)
	if err != nil {
		return err
	}

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		topic.ReplyTime = replies[0].Created
		topic.ReplyCount = int64(len(replies))
		_, err = o.Update(topic)
	}
	return err
}
