# syntax=docker/dockerfile:1

ARG KIND=plugins
ARG USERNAME=codeclarity
ARG PLUGINNAME=plugin
ARG GOLANG_VERSION=1.23.1
ARG PLATFORM=arm64

FROM r-base:4.4.1 AS plugin-base
ARG GOLANG_VERSION
ARG PLATFORM
# Install system dependencies
RUN apt-get update
RUN apt-get install git -y curl libhdf5-dev libxml2-dev libfribidi-dev libpng-dev libtiff5-dev libjpeg-dev libgit2-dev r-cran-devtools

# Install R packages
RUN R -e "install.packages(c('Seurat', 'hdf5r', 'tidyverse', 'ggplot2', 'svglite', 'rjson', 'future'))"

# Install Go
RUN curl -fsSL https://golang.org/dl/go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz -o go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz
RUN tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz
ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin

ADD .cloud/docker/config/* /root

# DEV IMAGE
FROM plugin-base AS plugin-dev
ARG KIND
ARG PLUGINNAME
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}

RUN go install github.com/air-verse/air@latest
CMD air -c .air.toml

# DEBUG IMAGE
FROM plugin-base AS plugin-debug
ARG KIND
ARG PLUGINNAME
ENV KIND ${KIND}
ENV PLUGINNAME ${PLUGINNAME}

# Install Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest
ADD .cloud/docker/config/* /root
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
CMD dlv debug . --headless --listen=:40000 --accept-multiclient
# CMD tail -f /dev/null


# BUILD IMAGE
FROM golang:1.23.0 AS plugin-build
ARG KIND
ARG PLUGINNAME
COPY . /codeclarity/${KIND}/${PLUGINNAME}
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN rm go.work
# RUN apt-get update
# RUN apt-get install -y git
RUN [ "go", "build", "-o", "plugin" ]

# PRODUCTION IMAGE
FROM r-base:4.4.1 AS plugin
ARG KIND
ARG USERNAME
ARG PLUGINNAME
ARG GOLANG_VERSION
ARG PLATFORM

RUN useradd -m ${USERNAME}
RUN mkdir /private
RUN chown -R ${USERNAME}:${USERNAME} /private
# Install system dependencies
RUN apt-get update
RUN apt-get install -y git curl libhdf5-dev libxml2-dev libfribidi-dev libpng-dev libtiff5-dev libjpeg-dev libgit2-dev r-cran-devtools

USER ${USERNAME}
RUN mkdir /home/${USERNAME}/R_LIBS
ENV R_LIBS_USER=/home/${USERNAME}/R_LIBS
# Install R packages
RUN R -e "install.packages(c('Seurat', 'hdf5r', 'tidyverse', 'ggplot2', 'svglite', 'rjson'))"

WORKDIR /home/${USERNAME}
# Install Go
RUN curl -fsSL https://golang.org/dl/go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz -o go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz
RUN tar -C /home/${USERNAME} -xzf go${GOLANG_VERSION}.linux-${PLATFORM}.tar.gz
ENV PATH=$PATH:/home/${USERNAME}/go/bin

WORKDIR /home/${USERNAME}/${PLUGINNAME}
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/plugin .
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/config.json .

ENTRYPOINT [ "./plugin" ]
