package go2cache

import "fmt"



type Region struct {
	Name string //当前region名称
	Size int    //当前区域下 对应cache的容积大小
}

func (r *Region) String() string {
	return fmt.Sprintf("region:%s@%d",r.Name,r.Size)
}
