export NS=kube2iam-test
export ROLE_NAME=kube2iam-${USER}-test

cat kube2iam-test-pod.yaml | envsubst | kubectl -n ${NS} apply -f -

while [[ $(kubectl -n ${NS} get pod kube2iam-test-pod -o jsonpath='{.status.phase}' | grep Pending )	 ]] ; do sleep 2 ; done

#kubectl -n ${NS} log kube2iam-test-pod

if [[ $(kubectl -n ${NS} get pod kube2iam-test-pod -o=jsonpath='{.status.phase}' | grep Failed ) ]] 
then
	echo "Failed"
	if [[ $(kubectl -n ${NS} logs kube2iam-test-pod | grep InvalidClientTokenId ) ]] 
	then 
		echo "The iam role has probably not been created yet. try again in few minutes"
	fi
else 
	echo "Success" 
fi
kubectl -n ${NS} delete pod kube2iam-test-pod