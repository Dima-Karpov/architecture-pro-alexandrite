' Task4: синий = источники логов и ELK
AddElementTag("existing", $bgColor="#F4F6F7", $fontColor="#1C2833", $borderColor="#7F8C8D", $legendText="Уже есть / контекст")
AddElementTag("logs", $bgColor="#D6EAF8", $fontColor="#1B4F72", $borderColor="#2874A6", $legendText="Источник логов / ELK")
AddRelTag("existing", $textColor="#566573", $lineColor="#7F8C8D", $legendText="Связь as-is / to-be")
AddRelTag("logs", $textColor="#1A5276", $lineColor="#2874A6", $legendText="Ship / search логов")
