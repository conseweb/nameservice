# Farmer APIs

## Account

- Post `/signup/{phone|email}`
- Post `/signup/`

### 注册

注册流程：

#### 1. 注册邮箱或手机

POST `/signup/{phone|email}`

body:
```
{
	"email": "me@ckeyer.com"
}
```
或者

```
{
	"phone": "18012345678"
}
```

返回 `200` 成功,
其它错误为（下面其它接口的错误返回结构与此一致）：
```
{
	"error":"错误信息",
	"message":"其它补充信息"
}
```

#### 2. 补全其它信息并注册用户

Post `/signup/`

body:
```
{
	"email": "me@ckeyer.com",
	"captcha": "2z6yd1", 
	"language": "简体中文",
	"nickname": "hello",
	"password": "hellohello",
	"type": "email" // 注册类型， `phone` 或者 `email`
}

```

ret: 
```
{"id":"6e8bbb55-aa46-11e6-8318-0242ac100103",
	"nicename":"hello",
	"phone":"",
	"email":"me@ckeyer.com",
	"password":"hellohello",
	"lang":1,
	"passphrase":"阀 得 惯 圈 睡 罗 售 推 习 驻 呵 阔 丹 壁 拆 热 昆 邀 写 格 阔 仪 淡 议",
	"Wallet":{},
	"devices":[{
		"userID":"6e8bbb55-aa46-11e6-8318-0242ac100103",
		"deviceID":"171709d7800",
		"os":"linux_amd64",
		"mac":"02:42:ac:10:01:05",
		"alias":"linux_amd64",
		"wpub":"dHB1YkRBN2NNWVlzWGpDV1NZeGl2UGdBbnhTVW1LN2dQQjRvYks3MW1tVHBCUlRyd01mZzZLTW84U0w1UHNtWnhDbmVqbVVvM1FzcUd4RlB1ZnhoejNSRjRqa1JGNGtXbU1hbmplZk5uZkZQUmg1",
		"spub":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAERJapdHzE6Ct27UouODxe7OfuUndLeamB/m1CRcX6O8lnOIpt7ofomJ9Ms4cKyLVjnV9izMYa0/4+sD2hgGdFLw==",
		"IsLocal":true,
		"Wallet":{},
		"address":"moKPkr8aw8ZDnW4oG8AH5UBpuqht1YRypi"
	}]
}
```


### 登录

POST `/account/login`

body
```
{
	"password": "hellohello",
	"email": "wcj0256@foxmail.com"
}
```

return `200`: 
```
{
	"id":"6e8bbb55-aa46-11e6-8318-0242ac100103",
	"nicename":"hello",
	"phone":"",
	"email":"wcj0256@foxmail.com",
	"password":"",
	"lang":0,
	"passphrase":"",
	"Wallet":null,
	"devices":[{
		"userID":"6e8bbb55-aa46-11e6-8318-0242ac100103",
		"deviceID":"171709d7800",
		"os":"linux_amd64",
		"mac":"02:42:ac:10:01:05",
		"alias":"linux_amd64",
		"wpub":"dHB1YkRBN2NNWVlzWGpDV1NZeGl2UGdBbnhTVW1LN2dQQjRvYks3MW1tVHBCUlRyd01mZzZLTW84U0w1UHNtWnhDbmVqbVVvM1FzcUd4RlB1ZnhoejNSRjRqa1JGNGtXbU1hbmplZk5uZkZQUmg1",
		"spub":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAERJapdHzE6Ct27UouODxe7OfuUndLeamB/m1CRcX6O8lnOIpt7ofomJ9Ms4cKyLVjnV9izMYa0/4+sD2hgGdFLw==",
		"IsLocal":true,
		"Wallet":null,
		"address":"moKPkr8aw8ZDnW4oG8AH5UBpuqht1YRypi"
	}]
}
```

### 退出
DELETE `/account/logout`

## 设备

### 获取交易序列

POST `/device/tx`

body:
```
{
	"in": [{
		"addr": "moKPkr8aw8ZDnW4oG8AH5UBpuqht1YRypi",
		"balance": 100000000,
		"pre_tx_hash": "c7afc4d32a03584ddd403ebd0145a2fed4d0b40005a61e10976ef403b6b483f6",
		"tx_out_index": "0"
	}],
	"out": [{
		"addr": "asdf",
		"amount": 10
	}]
}
```

return:

```
{
	"message":"CAEQ8YmmwQUaZhJAYzdhZmM0ZDMyYTAzNTg0ZGRkNDAzZWJkMDE0NWEyZmVkNGQwYjQwMDA1YTYxZTEwOTc2ZWY0MDNiNmI0ODNmNhoibW9LUGtyOGF3OFpEblc0b0c4QUg1VUJwdXFodDFZUnlwaSIICAoSBGFzZGYiKQj2wdcvEiJtb0tQa3I4YXc4WkRuVzRvRzhBSDVVQnB1cWh0MVlSeXBpKiJtb0tQa3I4YXc4WkRuVzRvRzhBSDVVQnB1cWh0MVlSeXBp"
}
```


## Peer控制

### 启动/停止
PATCH `/peer/{start|stop|restart}`