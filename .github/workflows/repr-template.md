### checkout this branch
```
mkdir -p $HOME/reflect/{branch_name} && cd $HOME/reflect/{branch_name}
git init && git remote add origin git@github.com:ok-john/reflect.git
git fetch origin --tags && git fetch origin pull/{pull_id}/head:{branch_name}
git checkout {branch_name}
``` 
