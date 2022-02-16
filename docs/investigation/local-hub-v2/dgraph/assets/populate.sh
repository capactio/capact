# Retry a command
n=0
until [ "$n" -ge 5 ]
do
   curl -X POST http://dgraph:8080/admin/schema --data-binary '@schema.graphql' && break  # substitute your command here
   n=$((n+1))
   sleep 5
done
