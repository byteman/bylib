package wifi

import "testing"

func TestGetApConnState(t *testing.T) {
	var apcli0=`apcli0    ESSID: "OpenWrt"
          Access Point: 00:0C:43:76:20:20
          Mode: Client  Channel: 11 (unknown)
          Tx-Power: unknown  Link Quality: unknown/100
          Signal: unknown  Noise: unknown
          Bit Rate: 72.0 MBit/s
          Encryption: unknown
          Type: wext  HW Mode(s): unknown
          Hardware: unknown [Generic WEXT]
          TX power offset: unknown
          Frequency offset: unknown
          Supports VAPs: no  PHY name: apcli0`
	lines,err:=ReadLine(apcli0)
	if err!=nil{
		t.Errorf("err=%s\n",err)
		return
	}
	t.Log("lines=",len(lines))
	for _,line:=range lines{
		t.Log(line)
	}
	var state ApConnState
	err=parseApState(lines,&state)

	t.Log(err)
	if err==nil{
		t.Logf("%+v",state)
	}
}
