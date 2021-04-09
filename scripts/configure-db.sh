# In the real world this script wouldn't exist and as this database would be configured manually for optimal results
set -m

/entrypoint.sh couchbase-server &

# UNSAFE!! production use would use a vault/credential keystore manager
COUCHBASE_ADMINISTRATOR_USERNAME=admin
COUCHBASE_ADMINISTRATOR_PASSWORD=password
COUCHBASE_BUCKET=messageBox
CLUSTER_NAME=cluster01

sleep 50 # The long sleep is needed as the DB can have a long startup getting ready before any init call

# Setup initial cluster/ Initialize Node
couchbase-cli cluster-init -c 127.0.0.1 --cluster-name $CLUSTER_NAME --cluster-username $COUCHBASE_ADMINISTRATOR_USERNAME \
--cluster-password $COUCHBASE_ADMINISTRATOR_PASSWORD --services data,index,query,fts --cluster-ramsize 256 --cluster-index-ramsize 256 \
--cluster-fts-ramsize 256 --index-storage-setting default \

# Setup Administrator username and password
curl -v http://127.0.0.1:8091/settings/web -d port=8091 -d username=$COUCHBASE_ADMINISTRATOR_USERNAME -d password=$COUCHBASE_ADMINISTRATOR_PASSWORD

sleep 15

# Setup Bucket
couchbase-cli bucket-create -c 127.0.0.1:8091 --username $COUCHBASE_ADMINISTRATOR_USERNAME \
--password $COUCHBASE_ADMINISTRATOR_PASSWORD  --bucket $COUCHBASE_BUCKET --bucket-type couchbase \
--bucket-ramsize 256

sleep 15

fg 1