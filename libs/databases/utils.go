package databases

import (
	"xorm.io/xorm"
)

func CheckWhereInfo[T any](w string, v *T, tp string, sess *xorm.Session) {
	if v != nil {
		//sess.Where(fmt.Sprintf(w, *v))
		switch tp {
		case "and":
			sess.And(w, *v)
		case "in":
			sess.In(w, *v)
		case "or":
			sess.Or(w, *v)
		case "notin":
			sess.NotIn(w, *v)
		case "where":
			sess.Where(w, *v)
		default:
			return
		}
	}
}
