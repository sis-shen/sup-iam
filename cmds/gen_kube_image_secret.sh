kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=你的GitHub用户名 \
  --docker-password=你的GitHub Personal Access Token \
  --docker-email=你的邮箱