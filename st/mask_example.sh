echo "creating mask block"
curl "localhost:8888/create?blockType=mask&id=m_mask"
echo "setting rule for mask"
curl "localhost:8888/blocks/m_mask/set_rule" --data '{"Mask":{"a":{},"c":{"d":{}}}}'
echo "creating random bock"
curl "localhost:8888/create?blockType=random&id=m_random"
echo "connecting random to mask"
curl "localhost:8888/connect?from=m_random&to=m_mask"
echo "creating tolog block"
curl "localhost:8888/create?blockType=tolog&id=m_log"
echo "connecting mask to log"
curl "localhost:8888/connect?from=m_mask&to=m_log"
