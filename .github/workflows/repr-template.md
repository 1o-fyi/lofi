### {branch_name} | checkout this branch
```
mkdir -p $HOME/1o-fyi/lofi/{branch_name} && cd $HOME/lofi/1o-fyi/{branch_name}
git init
git remote add origin https://github.com/1o-fyi/lofi.git
git fetch origin --tags && git fetch origin pull/{pull_id}/head:{branch_name}
git checkout {branch_name}
``` 
