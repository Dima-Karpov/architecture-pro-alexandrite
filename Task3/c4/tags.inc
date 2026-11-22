' Task3: не путать с тегом new (зелёный) из Task1
AddElementTag("existing", $bgColor="#F4F6F7", $fontColor="#1C2833", $borderColor="#7F8C8D", $legendText="Уже есть / контекст")
AddElementTag("trace", $bgColor="#FADBD8", $fontColor="#7B241C", $borderColor="#C0392B", $legendText="Трейсинг (новое)")
AddElementTag("alert", $bgColor="#D5F5E3", $fontColor="#145A32", $borderColor="#1E8449", $legendText="Алертинг по трейсам")
AddRelTag("existing", $textColor="#566573", $lineColor="#7F8C8D", $legendText="Связь as-is / to-be")
AddRelTag("trace", $textColor="#922B21", $lineColor="#C0392B", $legendText="OTLP / propagation")
AddRelTag("alert", $textColor="#196F3D", $lineColor="#1E8449", $legendText="Spanmetrics → алерт")
