# FROM ubuntu
# RUN apt-get update && apt-get -y install sudo vim

FROM centos
RUN yum install -y sudo

ADD users /users
ADD cruser /usr/local/bin/cruser
ADD docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod 755 /usr/local/bin/docker-entrypoint.sh
RUN echo "ssh-rsa asdasdasddasd test@asdasd.com" > /tmp/init_user && cruser -file /tmp/init_user
RUN echo "ssh-dss erererererere test@rerere.com" > /tmp/init_user && cruser -file /tmp/init_user

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
