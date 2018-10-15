**docker2apache**

Small glue utility to generate proxyPass-statements for Apache or other webservers by scanning a 
docker daemon for a target configuration that is tagged by docker labels.  

**Usage**

1.) Start your containers or a stack and attach docker labels with a reversed domain name and a vhost or uri_maps 
parameter set, e.g. 

labels:
- "com.example.vhost=devtest.example.com"             
- "com.example.uri_maps=/otrs:/otrs,/web1:/web1server"
`       

2.) Run  docker2apache and set the targeted docker daemon by an environment variable. As a parameter you need to pass
the reversed domain name, the  vhost and the private port for your service. Docker2apache respects all docker client 
environment variables.
 
`
DOCKER_HOST=http://192.168.5.144:2375 docker2apache -d com.example  -v devtest.example.com -p 8443 -x https`


3.) Pipe the output in a file, that is e.g. included in a full Apache Vhost configuration file.

` DOCKER_HOST=http://192.168.5.144:2375 docker2apache -d com.example  
-v devtest.example.com -p 8443 -x https > /etc/httpd/conf.d/devtest.docker.conf  && systemctl reload httpd`



   