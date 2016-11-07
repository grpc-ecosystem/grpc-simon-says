# Copyright 2015 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM grpc/go:1.0

RUN apt-get update && \
    apt-get -y install wget curl less unzip python zsh git make openssh-server libapparmor1 nano

WORKDIR /

#sshd setup - https://docs.docker.com/examples/running_ssh_service/
RUN mkdir /var/run/sshd
RUN echo 'root:pw' | chpasswd
RUN sed -i 's/PermitRootLogin without-password/PermitRootLogin yes/' /etc/ssh/sshd_config
# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd
ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile
EXPOSE 22

#install cloud sdk
RUN wget -q https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.zip && unzip -q google-cloud-sdk.zip && rm google-cloud-sdk.zip
RUN /google-cloud-sdk/install.sh --usage-reporting=true --path-update=true --bash-completion=true --rc-path=/root/.bashrc
ENV PATH /google-cloud-sdk/bin:$PATH

RUN gcloud components update kubectl --quiet

#redis
RUN mkdir redis && cd redis && wget -O redis.tar.gz -q http://download.redis.io/releases/redis-3.0.6.tar.gz && \
    tar xzf redis.tar.gz --strip-components=1 && rm redis.tar.gz && make
RUN sed -i "s/daemonize no/daemonize yes/" /redis/redis.conf
RUN sed -i 's|logfile ""|logfile "/redis/redis.log"|' /redis/redis.conf

RUN ln -s /redis/src/redis-cli /usr/local/bin/redis-cli

#oh-my-zsh, because how do we live without it?
RUN git clone https://github.com/robbyrussell/oh-my-zsh.git

#tools for working with go.
RUN go get github.com/constabulary/gb/...
RUN go get github.com/golang/lint/golint
RUN go get golang.org/x/tools/cmd/goimports
RUN go get -u github.com/kisielk/errcheck

#This is for protobuf
RUN go get -u github.com/golang/protobuf/protoc-gen-go

WORKDIR /root

#install kubectl visualiser
RUN git clone https://github.com/saturnism/gcp-live-k8s-visualizer.git
EXPOSE 8001

ADD startup.sh startup.sh
RUN chmod +x startup.sh

EXPOSE 7080-7090
EXPOSE 50051
