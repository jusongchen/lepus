// copyright Jusong Chen

package main

import (
	"flag"
	"fmt"
	// "database/sql"
	sql "github.com/jmoiron/sqlx"
	"os"
	"path/filepath"

	"github.com/jusongchen/lepus/version"

	"github.com/jusongchen/lepus/app"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

const defaultPort = "8080"
const receiveDir = "received"
const dirForStatic = "public"
const imageRelativeDir = "images"
const viewPath = "views"
const sqliteFile = "lepus.v1.DB"
const sessionKeyEnvName = "LEPUS_SESSION_KEY"

func main() {
	log.Infof("Lepus version:%s", version.Release)
	var err error

	lepusHomeDir := os.Getenv("LEPUS_HOME")
	if lepusHomeDir == "" {

		lepusHomeDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("Lepus home dir:%s", lepusHomeDir)

	if err := os.Chdir(lepusHomeDir); err != nil {

		log.WithError(err).Fatalf("Chdir to %v failed", lepusHomeDir)
	}

	if _, err := os.Stat(dirForStatic); os.IsNotExist(err) {
		log.Fatalf("Directory for static web content does not exist:%s", dirForStatic)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s -port <port number>\n", os.Args[0])
		flag.PrintDefaults()
	}

	portInt := flag.Int("port", 8080, "TCP port to listen on")
	flag.Parse()

	sessionKey := os.Getenv(sessionKeyEnvName)
	if len(sessionKey) == 0 {
		log.Fatalf("Environment Variable %v not found.", sessionKeyEnvName)
	}

	educatorNames := []string{"蔡春耀", "蔡温榄", "蔡小华", "曹世才", "曾敏毅", "沈淑耀", "陈本培",
		"陈福音", "陈济兴", "陈金治", "陈群青", "陈细英", "陈燕华", "陈由溪", "陈玉珍", "陈长青", "陈志章",
		"陈宗辉", "傅德卿", "官尚武", "何天祥", "洪大锋", "胡永模", "胡永岳", "江英岩", "蒋永潮", "李粹玉",
		"练锡康", "林秉松", "林端川", "林芳", "林嘉坚", "林茂英", "林培坚", "林岫英", "林月心", "林占樟",
		"林昭英", "刘开天", "刘世煌", "刘永宾", "陆佩珰", "罗朝东", "毛玉珍", "毛祖辉", "毛祖瑜", "欧阳兴",
		"潘宝升", "潘家健", "潘世英", "潘孝平", "钱振恒", "石柏仁", "童美霞", "王福庆", "王光琳", "王丽华",
		"王美珠", "王其本", "王强", "王如", "王玉芳", "危金炎", "魏友义", "吴大樑", "吴齐练", "吴生基",
		"肖方尤", "肖光磊", "肖忠波", "许美钗", "严秀凤", "杨诚", "杨虹", "杨孝华", "杨义森", "叶菁", "余春华",
		"余冬生", "余世棣", "余天雨", "余伟然", "詹玉赐", "张殿", "张桂贞", "张家新", "张孔烺", "张荣治",
		"张世年", "张芝萱", "赵文婷", "郑春高", "郑国钦", "郑玉珠", "郑宗振", "周治河", "庄可明", "庄瑞发",
		"陈松赞", "张延勃", "张烺官", "陈岚", "游梅英", "张文", "郑秀峰", "陈兆华", "吴建芳", "王爱玲", "陈本端",
		"陈建尧", "陈振城", "张庆明", "黄光全", "张丽贞", "林成键", "阮珠珍", "赖汾扬", "林为炎", "魏忠麟",
		"陈冠民", "林惠琛"}

	db, err := sql.Open("sqlite3", filepath.Join(receiveDir, sqliteFile))
	if err != nil {
		log.Fatalf("Faile top open sqlite3 DB:%v", err)
	}
	log.Infof("Lepus DB file:%s", sqliteFile)

	app.Start(db,
		sessionKey,
		fmt.Sprintf(":%d", *portInt),
		lepusHomeDir,
		dirForStatic,
		receiveDir,
		imageRelativeDir,
		viewPath,
		educatorNames)

}
