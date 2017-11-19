if [ -z "$TRANSFORM" ]; then
    echo "No $TRANSFORM path"
    exit 1
fi

if [ -z "$URL" ]; then
    echo "No $URL"
    exit 1
fi

rm -rf ${TRANSFORM}
rm -rf wget
mkdir wget/
for i in `seq 1 100`; do
    wget  ${URL} -P wget -O wget/img.$i.jpg  &
done

rm -rf ${TRANSFORM}
