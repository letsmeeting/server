## **lilpop-server : Lilpop Server Application**

### Features

- [x] Backend API Server (RESTful API)
- [x] Application Management (Run/Stop)
- [x] Database Management (MySQL, redis, mongodb)
- [x] Logging

***

### Prerequisite

- go 언어 설치 및 개발환경 설정

```shell
# Go 최신버전 다운로드 및 설치
https://golang.org/dl/
$ export GOLANG_VERSION=1.16.3 # 버전 선택
$ sudo tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz # 압축해제 /usr/local/go
# GOROOT, GOPATH 설정
$ vi ~/.bashrc # vi로 편집하여 추가
  export GOROOT=/usr/local/go        # Go 설치경로
  export GOPATH=${HOME}/work/go   # 개인 Go 개발환경
  export GOPRIVATE=github.com/jinuopti/* # set private git repo
  export POSGO_PATH=${GOPATH}/src/github.com/jinuopti/posgo  # posGo 경로 필수 설정!
  PATH=$PATH:$GOROOT/bin:$GOPATH/bin # GOPATH/bin PATH 추가

# bin pkg src 기본 설치  
$ go get golang.org/x/tools/cmd/...

# go module 오류 방지를 위해 bitbucket ssh 대신 personal access token 사용 권장
# 생성 : Bitbucket - Manage Account - Personal access tokens - Create a token - token 복사
export PA_TOKEN={복사한 token}  # .bashrc 에 추가
# htts personal access token 추가
$ git config --global url."https://${PA_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"
# 이후 ssh 대신 https 사용  
```

### Build & Run Application

```shell
$ git clone git@github.com:jinuopti/lilpop-server.git
$ cd lilpop-server
lilpop-server$ ./build.sh # lilpop binary 생성
lilpop-server$ cp lilpop.ini.sample lilpop.ini  # sample 설정파일 복사
lilpop-server$ vi lilpop.ini # 환경설정 파일 수정 (경로 등)
lilpop-server$ ./lilpop -h # help message
lilpop-server$ ./lilpop # start application
# pm-engine.ini 의 [LOG] -> LogFile 경로에 Log 파일 생성
lilpop-server$ cat log/lilpop.log
```
# server
