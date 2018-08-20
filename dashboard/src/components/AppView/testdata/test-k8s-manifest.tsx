const manifest = `
apiVersion: extensions/v1beta1
kind: Deployment  
metadata:                    
  annotations:                  
    deployment.kubernetes.io/revision: "1"
  generation: 1            
  labels:            
    app: redis                                      
    chart: redis-3.7.6                
  name: deployment-one
spec:                            
  replicas: 1
  revisionHistoryLimit: 10
---
apiVersion: v1
kind: Service
metadata:                    
  labels:            
    app: redis                                      
    chart: redis-3.7.6                
  name: svc-one
spec:                            
  clusterIP: 100.70.47.47
---
apiVersion: v1
kind: ConfigMap
metadata:                    
  labels:            
    app: redis                                      
    chart: redis-3.7.6                
  name: cm-one
spec:                            
  clusterIP: 100.70.47.47
---
apiVersion: v1
metadata:                    
  labels:            
    app: redis                                      
    chart: redis-3.7.6                
  name: broken
---
`;

export default manifest;
