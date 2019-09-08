# disney-bookinger

## 日毎のデータ(POST)
https://reserve.tokyodisneyresort.jp/sp/showRestaurant/ajaxReservationOfDate/

_xhr=&commodityCD=XSS05RDHS0023&contentsCd=05&nameCd=RDHS0&showId=3&useDate=20190905&adultNum=1&childNum=0&childAgeInform=&wheelchairCount=0&stretcherCount=0&reservationStatus=0&seatRank=A

### params
* commodityCD
商品コード、ザ・ダイヤモンドホースシュー・プレゼンツ“ミッキー＆カンパニー”は
`XSS05RDHS002[2-4]`、 ランクとIDの紐付けは `2: S`, `3: A`, `4: B`
* nameCd
(おそらく)レストランのコード
* showId
わからん
* useData
空席確認する日付(yyyymmdd形式)
* adultNum, shildNum
読んで字のごとく

## 週毎のデータ(POST)
https://reserve.tokyodisneyresort.jp/sp/showRestaurant/ajaxReservationOfWeek/

_xhr=&useDate=20190903&showId=3&contentsCd=05&nameCd=RDHS0&adultNum=1&childNum=0&childAgeInform=&wheelchairCount=0&stretcherCount=0

## 空席状況
満席/空席あり のどちらかで表示

## 予約確定エンドポイント
https://reserve.tokyodisneyresort.jp/common/notice/

### Params(POST)
```
showId: 1
contentsCd: 05
nameCd: RDHS0
materialCd: RLDHS1B004
openNumKey: 1
commodityCD: XXXRLDHS1B004
useDate: 20190906
adultNum: 1
childNum: 0
searchChildAges: 
wheelChairCount: 0
stretcherCount: 0
dispExhibitionTimes: 1
exhibitionTime: 1040
seatRank: B
reservationStatus: 0
org.apache.struts.taglib.html.TOKEN: IAAAOI63COCG227MOHFYOZ0Z0FJHCQVK
prev: /showrestaurant/search/?useDate=20190906&adultNum=1&childNum=0&childAgeInform=&wheelchairCount=0&stretcherCount=0&freeword=&reservationStatus=0
```