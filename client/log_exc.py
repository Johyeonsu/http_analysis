import json
import sys

all_data=[]
start = 0
end = 0
handshakedone = 0
firstpacket = 0
fin = 0

with open(sys.argv[1], 'r') as f:
    for line in f:
        json_data = json.loads(line)
        all_data.append(json_data)


for i, data in enumerate(all_data):
    if i == 0:
        continue
    else:
        if data["name"] == "transport:connection_started":
            start = data["time"]

        if data["name"] == "transport:connection_closed":
            end = data["time"]

        if (data["name"] == "transport:packet_sent" and
            data["data"]["header"]["packet_type"] == "initial" and
            data["data"]["header"]["packet_number"] ==  0):
            firstpacket = data["time"]

        if "frames" in data["data"] :
            for frame in data["data"]["frames"]:
                if frame["frame_type"] == "handshake_done":
                    handshakedone = data["time"]
                if "fin" in frame and frame["fin"] == True:
                    fin = data["time"]
                if "stream_type" in frame and frame["stream_type"] == "bidirectional":
                    end = data["time"]
                    
p = "%-20s%20f\n%-20s%20f\n%-20s%20f\n%-20s%20f\n%-20s%20f"%("Connect Start :", start,"Initial Packet :", firstpacket,"Handshake Done :", handshakedone, "Finish :", fin, "Connection Close :", end)
# print(p, file=open("tmp", 'w'))
print(p)
