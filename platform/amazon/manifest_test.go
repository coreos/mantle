package amazon

import (
	"encoding/xml"
	"testing"
)

const testinput = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<manifest>
    <version>2010-11-15</version>
    <file-format>VMDK</file-format>
    <importer>
        <name>ec2-upload-disk-image</name>
        <version>1.0.0</version>
        <release>2010-11-15</release>
    </importer>
    <self-destruct-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdkmanifest.xml?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=IlJW8yHFIvzYY9N4IsYE82%2Fv7wQ%3D</self-destruct-url>
    <import>
        <size>158396928</size>
        <volume-size>9</volume-size>
        <parts count="16">
            <part index="0">
                <byte-range end="10485759" start="0"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part0</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part0?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=r3RI7%2BhVk%2BnAkub2oBtrES4AMfY%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part0?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=Hklb4bq6nxPFp29%2FhQzVywsZcVQ%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part0?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=NDZv6UhIAVU%2FApn%2FGJaHwBYUkAY%3D</delete-url>
            </part>
            <part index="1">
                <byte-range end="20971519" start="10485760"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part1</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part1?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=gNVyTmuqfNkq7rFF0T3TRtqwZzU%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part1?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=4J3L9VcDJcO3kTmwZwc92%2F3iy0I%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part1?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=%2Bxa9Yh%2BUBDG3EsatXjlKuvS5hNE%3D</delete-url>
            </part>
            <part index="2">
                <byte-range end="31457279" start="20971520"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part2</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part2?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=LmnU77r92pSJ3mTiwrgch3DxY3w%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part2?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=RuNNDsuXG2SnrHsAhqd1N%2B0FAhQ%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part2?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=nw0aGlE8guPVWHxP1bfmKrdv1zI%3D</delete-url>
            </part>
            <part index="3">
                <byte-range end="41943039" start="31457280"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part3</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part3?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=aVvrtgdsVlwkKYjIhyGOu6Fu5og%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part3?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=szoVF9%2FGkvUOTPXtO%2BoIzsTBkeA%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part3?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=7MNRBADEdFZwwrXRMZjP3zcldvg%3D</delete-url>
            </part>
            <part index="4">
                <byte-range end="52428799" start="41943040"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part4</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part4?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=JNapyk7yLnnx%2FLZLocMMjYskQ6U%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part4?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=Jq9OcJCpYVpsPxQnI7aO1saXkfk%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part4?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=z2BXH4X0i0m70gX0c0YdCi4UCRg%3D</delete-url>
            </part>
            <part index="5">
                <byte-range end="62914559" start="52428800"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part5</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part5?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=ZxyctZV%2FfkfmKGQJe1M87rzaUdo%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part5?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=cxSDWc2Z7OmjbsFN4yR2hfDdAZo%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part5?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=MxQXvNsZllUU5YKmtf5%2FvVJWeUA%3D</delete-url>
            </part>
            <part index="6">
                <byte-range end="73400319" start="62914560"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part6</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part6?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=VogWVj36ydo8FcbelbxeEPypUkM%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part6?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=XrZiABCBnWZSrJBKgN%2Fn0cFXS%2BU%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part6?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=giz8eFPd%2FDwFEkLlQKLrWKbDoPw%3D</delete-url>
            </part>
            <part index="7">
                <byte-range end="83886079" start="73400320"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part7</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part7?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=GGxwlT8frMN%2FA7CITIlPirRA1Xc%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part7?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=lkxLBYDz%2FAe%2FYrGddKTcsgmjq%2FI%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part7?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=TEoXXK2pmWWQX1%2FxClL5%2FV3G92A%3D</delete-url>
            </part>
            <part index="8">
                <byte-range end="94371839" start="83886080"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part8</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part8?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=jLQ62ECGVq%2FXZ8a5tXGdrxLQsbU%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part8?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=68hsXjIqrwUXsOhtNaGVlftY5tQ%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part8?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=C14U%2FM0oxiuqqfhJcGudqQOdGSU%3D</delete-url>
            </part>
            <part index="9">
                <byte-range end="104857599" start="94371840"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part9</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part9?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=hvXln7UBBKFQQRyC6FsUoCHW3g8%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part9?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=llbBvI6JKQgO09%2BW%2Bss5tUfz2Hw%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part9?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=aBnNX7spGBY0mF8vYX2dQjLD184%3D</delete-url>
            </part>
            <part index="10">
                <byte-range end="115343359" start="104857600"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part10</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part10?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=%2BLuEL3aVjhVsr%2BNMIbquUX201Ps%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part10?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=4ccT67qiVLjtsl9nqPHf8ni60%2FQ%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part10?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=mfns%2FaFsooAtF80KWq7sbtDHFKw%3D</delete-url>
            </part>
            <part index="11">
                <byte-range end="125829119" start="115343360"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part11</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part11?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=1kmmh2Zk3RP3EFOmWje0xcDXz6c%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part11?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=N565%2FFvynrrz%2Bl3NGtaDk%2FhYAiE%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part11?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=GMxiHKLmc9YnQqJR8aCoQHnVdqk%3D</delete-url>
            </part>
            <part index="12">
                <byte-range end="136314879" start="125829120"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part12</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part12?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=WmlxOMzlDokNmywL0raqVdIUv40%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part12?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=a1m4EF%2BcOHWbonzltuUAIShTksk%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part12?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=UUK84xXIYR6Njllk4y0OgZj3MSk%3D</delete-url>
            </part>
            <part index="13">
                <byte-range end="146800639" start="136314880"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part13</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part13?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=iPYYmiEXNR1NxXmvH0BT7Mxe1xo%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part13?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=6GS3DtRYbXGRahMrXUmF%2FcQZb6Y%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part13?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=IaABaCuiXlEa4389TomvGWD33Uk%3D</delete-url>
            </part>
            <part index="14">
                <byte-range end="157286399" start="146800640"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part14</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part14?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=ow8jNw3%2FIx24K2UXSS8MoWgToWU%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part14?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=15sWfUxLKuDNvEuq4d10zkPoU1Y%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part14?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=QKFYXihuWkjmbKbF0FtGNxYHkx0%3D</delete-url>
            </part>
            <part index="15">
                <byte-range end="158396927" start="157286400"/>
                <key>edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part15</key>
                <head-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part15?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=peTB%2FAPmOpDUpAY6shPX6CyLiIw%3D</head-url>
                <get-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part15?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=6qPMYm0%2BEQrpEygIIAqLIGuw6uE%3D</get-url>
                <delete-url>https://marineam-import-testing.s3.amazonaws.com/edc188a4-5628-4716-9de3-96c05930318f/coreos_production_vmware_ova-disk1.vmdk.part15?AWSAccessKeyId=AKIAJG7X3NK3KWDYRVVQ&amp;Expires=1429231554&amp;Signature=oRdo8TMvPmSwscDdDdq9P219aOk%3D</delete-url>
            </part>
        </parts>
    </import>
</manifest>
`

func TestUnmarshalManifest(t *testing.T) {
	var m Manifest
	err := xml.Unmarshal([]byte(testinput), &m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("manifest: %+v", m)
}
