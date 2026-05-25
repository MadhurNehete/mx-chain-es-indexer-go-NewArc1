IMAGE_NAME=elastic-container
DEFAULT_ES_VERSION=7.16.2
DEFAULT_ES_USERNAME=elastic
PROMETHEUS_CONTAINER_NAME=prometheus_container
GRAFANA_CONTAINER_NAME=grafana_container
GRAFANA_VERSION=10.0.3
PROMETHEUS_VERSION=v2.46.0
INDICES_LIST=("rating" "transactions" "blocks" "validators" "miniblocks" "rounds" "accounts" "accountshistory" "receipts" "scresults" "accountsesdt" "accountsesdthistory" "epochinfo" "scdeploys" "tokens" "tags" "logs" "delegators" "operations" "esdts" "values" "events" "executionresults" "drwa-denials" "drwa-identities" "drwa-holder-compliance" "drwa-attestations" "drwa-token-policies" "drwa-control-events" "mrv-anchored-proofs")

elastic_username() {
  if [ -n "${ELASTIC_USERNAME}" ]; then
    echo "${ELASTIC_USERNAME}"
    return
  fi

  echo "${DEFAULT_ES_USERNAME}"
}

require_elastic_password() {
  if [ -n "${ELASTIC_PASSWORD}" ]; then
    return
  fi

  echo "ELASTIC_PASSWORD must be set before starting or deleting secured Elasticsearch indices."
  echo "Example: export ELASTIC_PASSWORD='<strong-local-password>'"
  exit 1
}

elastic_curl() {
  NETRC_FILE=$(mktemp)
  cleanup_netrc() {
    rm -f "${NETRC_FILE}"
    trap - RETURN INT TERM
  }
  trap cleanup_netrc RETURN INT TERM

  chmod 600 "${NETRC_FILE}"
  printf "machine localhost login %s password %s\n" "${ES_USERNAME}" "${ELASTIC_PASSWORD}" > "${NETRC_FILE}"

  curl --netrc-file "${NETRC_FILE}" "$@"
  CURL_STATUS=$?
  cleanup_netrc

  return ${CURL_STATUS}
}


start() {
  ES_VERSION=$1
  if [ -z "${ES_VERSION}" ]; then
    ES_VERSION=${DEFAULT_ES_VERSION}
  fi
  require_elastic_password
  ES_USERNAME=$(elastic_username)

  docker pull docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  docker rm -f ${IMAGE_NAME} 2> /dev/null
  docker run -d --name "${IMAGE_NAME}" -p 9200:9200  -p 9300:9300 \
   -e "discovery.type=single-node" -e "xpack.security.enabled=true" -e "ELASTIC_PASSWORD=${ELASTIC_PASSWORD}" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
    docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}
  # Wait elastic cluster to start
  echo "Waiting Elasticsearch cluster to start..."
  for _ in $(seq 1 60); do
    if elastic_curl -fsS "http://localhost:9200" > /dev/null; then
      break
    fi
    sleep 1s
  done
  docker ps -a
}

stop() {
  docker stop "${IMAGE_NAME}"
}

delete() {
   require_elastic_password
   ES_USERNAME=$(elastic_username)

   for str in ${INDICES_LIST[@]}; do
      elastic_curl -XDELETE http://localhost:9200/$str-000001
      elastic_curl -XDELETE http://localhost:9200/$str
      elastic_curl -s -o /dev/null -w "%{http_code}" -X GET localhost:9200/_ilm/policy/$str-policy | grep -q 200 && elastic_curl -X DELETE localhost:9200/_ilm/policy/$str-policy
      echo
   done

  elastic_curl -XDELETE http://localhost:9200/_template/*
  echo
}


IMAGE_OPEN_SEARCH=open-container
DEFAULT_OPEN_SEARCH_VERSION=1.2.4

start_open_search() {
  OPEN_VERSION=$1
  if [ -z "${OPEN_VERSION}" ]; then
    OPEN_VERSION=${DEFAULT_OPEN_SEARCH_VERSION}
  fi

  docker pull opensearchproject/opensearch:${OPEN_VERSION}

  docker rm -f ${IMAGE_OPEN_SEARCH} 2> /dev/null
  docker run -d --name "${IMAGE_OPEN_SEARCH}" -p 9200:9200 -p 9600:9600 \
   -e "discovery.type=single-node" -e "plugins.security.disabled=true" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
   opensearchproject/opensearch:${OPEN_VERSION}

}

stop_open_search() {
  docker stop "${IMAGE_OPEN_SEARCH}"
}

start_prometheus_and_grafana() {
 docker rm -f ${PROMETHEUS_CONTAINER_NAME} 2> /dev/null
 docker rm -f ${GRAFANA_CONTAINER_NAME} 2> /dev/null

 PROMETHEUS_CONFIG_FOLDER=$(pwd)/prometheus
 docker run --network="host" --name "${PROMETHEUS_CONTAINER_NAME}" -d -p 9090:9090 -v "${PROMETHEUS_CONFIG_FOLDER}/prometheus.yml":/etc/prometheus/prometheus.yml prom/prometheus:${PROMETHEUS_VERSION}
 docker run --network="host" --name "${GRAFANA_CONTAINER_NAME}" -d -p 3000:3000  grafana/grafana:${GRAFANA_VERSION}
}

stop_prometheus_and_grafana() {
  docker stop "${PROMETHEUS_CONTAINER_NAME}"
  docker stop "${GRAFANA_CONTAINER_NAME}"
}

"$@"
