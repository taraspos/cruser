# FROM ubuntu
# RUN apt-get update && apt-get -y install sudo vim

FROM centos
RUN yum install -y sudo

ADD users /users
ADD cruser /usr/local/bin/cruser
ADD docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod 755 /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]