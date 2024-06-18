/Users/emmanuel.colussi/Documents/App/sonar/sonar-scanner-5.0.1.3006-macosx/bin/sonar-scanner \
  -Dsonar.projectKey=GoLC \
  -Dsonar.sources=. \
 -Dsonar.scm.provider=git \
  -Dsonar.branch.name=ver1.0.3 \
-Dsonar.go.coverage.reportPaths=coverage.out \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=squ_156226eb038153d9971e047fcbb36cba6438936f\
  -Dsonar.exclusions=config.json,dist,Release,Saves,Results,config_sample.json,createrelease.sh,imgs,.git
