package models

type Admin struct {
  UUID     string `json:"UUID"`     //全局唯一的索引, LedisDB Admin List 保存全局所有的 UUID 列表信息，LedisDB独立保存每个用户信息到一个 hash,名字为 {UUID}
  Username string `json:"username"` //用于保存用户的登录名,全局唯一
  Password string `json:"password"` //保存系统管理员 MD5 后的密码
  Email    string `json:"email"`    //
  Created  int64  `json:"created"`  //系统管理员创建时间
  Updated  int64  `json:"updated"`  //系统管理员信息更新时间
  Memo     string `json:"memo"`     //
}
