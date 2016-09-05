/*
Copyright Mojing Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package protos

import "regexp"

const (
	// ^[a-zA-Z0-9_.]+@[a-zA-Z0-9-]+[.a-zA-Z]+$
	regexp_email = "^[a-zA-Z0-9_.]+@[a-zA-Z0-9-]+[.a-zA-Z]+$"

	// ^(0|86|17951)?(13[0-9]|15[012356789]|17[0678]|18[0-9]|14[57])[0-9]{8}$
	regexp_mobile = "^(0|86|17951)?(13[0-9]|15[012356789]|17[0678]|18[0-9]|14[57])[0-9]{8}$"
)

// verify whether acquireCaptcha request is ok
func (req *AcquireCaptchaReq) Validate() bool {
	var matched bool
	var err error

	switch req.SignUpType {
	case SignUpType_EMAIL:
		matched, err = regexp.MatchString(regexp_email, req.SignUp)
	case SignUpType_MOBILE:
		matched, err = regexp.MatchString(regexp_mobile, req.SignUp)
	}

	if err != nil || !matched {
		return false
	}

	return true
}

// SetSignature set signature
func (req *RegisterUserReq) SetSignature(sign []byte) {
	req.Sign = sign
}

// GetSignature get signature
func (req *RegisterUserReq) GetSignature() []byte {
	return req.Sign
}

// SetSignature set signature
func (req *BindDeviceReq) SetSignature(sign []byte) {
	req.Sign = sign
}

// GetSignature get signature
func (req *BindDeviceReq) GetSignature() []byte {
	return req.Sign
}
