package logger

import "github.com/sirupsen/logrus"

func Setup() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
	})

	logrus.SetReportCaller(true)
}
