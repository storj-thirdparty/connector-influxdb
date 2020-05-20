## verify documentation is present. 
## cd is /workspace
HOME_MD="./README.assets/Home.md"
README_MD="./README.md"
if [ -f $HOME_MD ] && [ -f $README_MD ]; then
    echo "doco check PASSED"
    exit 0
fi
echo "doco check FAILED"
exit 1
