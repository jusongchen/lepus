// copyright Jusong Chen

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jusongchen/lepus/version"

	"github.com/jusongchen/lepus/app"
	"github.com/sirupsen/logrus"
)

const defaultPort = "8080"
const receiveDir = "received"
const dirForStatic = "public"
const imageRelativeDir = "images"
const viewPath = "views"

func main() {
	logrus.Infof("Lepus version:%s", version.Release)

	lepusHomeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Lepus home dir:%s", lepusHomeDir)

	if err := os.Chdir(lepusHomeDir); err != nil {

		logrus.WithError(err).Fatalf("Chdir to %v failed", lepusHomeDir)
	}

	if _, err := os.Stat(dirForStatic); os.IsNotExist(err) {
		logrus.Fatalf("Directory for static web content does not exist:%s", dirForStatic)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s -port <port number>\n", os.Args[0])
		flag.PrintDefaults()
	}

	portInt := flag.Int("port", 8080, "TCP port to listen on")
	flag.Parse()

	educatorNames := []string{"蔡春耀", "蔡温榄", "蔡小华", "曹世才", "曾敏毅", "沈淑耀", "陈本培", "陈福音", "陈济兴", "陈金治", "陈群青", "陈细英", "陈燕华", "陈由溪", "陈玉珍", "陈长青", "陈志章", "陈宗辉", "傅德卿", "官尚武", "何天祥", "洪大锋", "胡永模", "胡永岳", "江英岩", "蒋永潮", "李粹玉", "练锡康", "林秉松", "林端川", "林芳", "林嘉坚", "林茂英", "林培坚", "林岫英", "林月心", "林占樟", "林昭英", "刘开天", "刘世煌", "刘永宾", "陆佩珰", "罗朝东", "毛玉珍", "毛祖辉", "毛祖瑜", "欧阳兴", "潘宝升", "潘家健", "潘世英", "潘孝平", "钱振恒", "石柏仁", "童美霞", "王福庆", "王光琳", "王丽华", "王美珠", "王其本", "王强", "王如", "王玉芳", "危金炎", "魏友义", "吴大樑", "吴齐练", "吴生基", "肖方尤", "肖光磊", "肖忠波", "许美钗", "严秀凤", "杨诚", "杨虹", "杨孝华", "杨义森", "叶菁", "余春华", "余冬生", "余世棣", "余天雨", "余伟然", "詹玉赐", "张殿", "张桂贞", "张家新", "张孔烺", "张荣治", "张世年", "张芝萱", "赵文婷", "郑春高", "郑国钦", "郑玉珠", "郑宗振", "周治河", "庄可明", "庄瑞发"}

	app.Start(fmt.Sprintf(":%d", *portInt),
		lepusHomeDir,
		dirForStatic,
		receiveDir,
		imageRelativeDir,
		viewPath,
		educatorNames)

}
