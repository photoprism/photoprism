/*
This package contains PhotoPrism core functionality.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package photoprism

import "github.com/sirupsen/logrus"

var log *logrus.Logger

func init() {
	log = logrus.StandardLogger()
}
