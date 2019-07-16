'''
sudo pip install DMXEnttecPro
'''
'''
from DMXEnttecPro import Controller
dmx = Controller('/dev/ttyUSB0')

# or 
'''


from DMXEnttecPro import Controller
from DMXEnttecPro.utils import get_port_by_serial_number, get_port_by_product_id
#my_port = get_port_by_serial_number('6A3ROIWE')
my_port = get_port_by_product_id(24577)
dmx = Controller(my_port)

dmx.set_channel(2,0)
dmx.set_channel(3,0)
dmx.submit()
