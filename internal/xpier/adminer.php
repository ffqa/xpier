<?php
/** Adminer - Compact database management
* @link https://www.adminer.org/
* @author Jakub Vrana, https://www.vrana.cz/
* @copyright 2007 Jakub Vrana
* @license https://www.apache.org/licenses/LICENSE-2.0 Apache License, Version 2.0
* @license https://www.gnu.org/licenses/gpl-2.0.html GNU General Public License, version 2 (one or other)
* @version 6.0.0
*/namespace
Adminer;const
VERSION="6.0.0";error_reporting(24575);set_error_handler(function($Dc,$Fc){return!!preg_match('~^Undefined (array key|offset|index)~',$Fc);},E_WARNING|E_NOTICE);$gd=!preg_match('~^(unsafe_raw)?$~',ini_get("filter.default"));if($gd||ini_get("filter.default_flags")){foreach(array('_GET','_POST','_COOKIE','_SERVER')as$X){$Bj=filter_input_array(constant("INPUT$X"),FILTER_UNSAFE_RAW);if($Bj)$$X=$Bj;}}if(function_exists("mb_internal_encoding"))mb_internal_encoding("8bit");function
connection($f=null){return($f?:Db::$instance);}function
adminer(){return
Adminer::$instance;}function
driver(){return
Driver::$instance;}function
connect(){$Ab=adminer()->credentials();$J=Driver::connect($Ab[0],$Ab[1],$Ab[2]);return(is_object($J)?$J:null);}function
idf_unescape($u){if(!preg_match('~^[`\'"[]~',$u))return$u;$Oe=substr($u,-1);return
str_replace($Oe.$Oe,$Oe,substr($u,1,-1));}function
q($Q){return
connection()->quote($Q);}function
idx($ua,$x,$j=null){return($ua&&array_key_exists($x,$ua)?$ua[$x]:$j);}function
number($X){return
preg_replace('~[^0-9]+~','',$X);}function
int_type(){return'(tiny|small|medium|big)?int(eger|\d)?';}function
number_type(){return'(^('.int_type().'|decimal|numeric|real|(binary_|half_|scaled_)?float\d?|(binary_)?double( precision)?|(small)?money)$)';}function
remove_slashes(array$Tj,$gd=false){$J=array();foreach($Tj
as$x=>$X)$J[stripslashes($x)]=(is_array($X)?remove_slashes($X,$gd):($gd?$X:stripslashes($X)));return$J;}function
bracket_escape($u,$Ca=false){static$mj=array(':'=>':1',']'=>':2','['=>':3','"'=>':4','='=>':5');return
strtr($u,($Ca?array_flip($mj):$mj));}function
url_escape($Q){static$mj=array();if(!$mj){$mj=array(' '=>'+');foreach(str_split("\"'<>#%&+=?".ini_get("arg_separator.input"))as$Sa)$mj[$Sa]=sprintf('%%%02X',ord($Sa));for($s=0;$s<256;$s++){if($s<32||$s>126)$mj[chr($s)]=sprintf('%%%02X',$s);}}return
strtr((string)$Q,$mj);}function
min_version($Wj,$hf="",$f=null){$f=connection($f);$fi=$f->server_info;if($hf&&preg_match('~([\d.]+)-MariaDB~',$fi,$B)){$fi=$B[1];$Wj=$hf;}return$Wj&&version_compare($fi,$Wj)>=0;}function
charset(Db$e){return(min_version("5.5.3",0,$e)?"utf8mb4":"utf8");}function
ini_set($rg,$Y){return(function_exists('ini_set')?\ini_set($rg,$Y):false);}function
ini_bool($oe){$X=ini_get($oe);return(preg_match('~^(on|true|yes)$~i',$X)||(int)$X);}function
ini_bytes($oe){$X=ini_get($oe);switch(strtolower(substr($X,-1))){case'g':$X=(int)$X*1024;case'm':$X=(int)$X*1024;case'k':$X=(int)$X*1024;}return$X;}function
max_input_vars($K,$Dg){$kf=(int)ini_get("max_input_vars");return($kf?(int)floor(($kf-$Dg)/$K):0);}function
max_input_vars_error(){$oe="max_input_vars";return
sprintf('Maximum number of allowed fields exceeded. Please increase %s.',"<b>$oe = ".ini_get($oe)."</b>");}function
sid(){static$J;if($J===null)$J=(SID&&!($_COOKIE&&ini_bool("session.use_cookies")));return$J;}function
set_password($Vj,$N,$V,$Ug){$_SESSION["pwds"][$Vj][$N][$V]=($_COOKIE["adminer_key"]&&is_string($Ug)?array(encrypt_string($Ug,$_COOKIE["adminer_key"])):$Ug);}function
get_password(){$J=get_session("pwds");if(is_array($J))$J=($_COOKIE["adminer_key"]?decrypt_string($J[0],$_COOKIE["adminer_key"]):false);return$J;}function
get_val($H,$l=0,$pb=null){$pb=connection($pb);$I=$pb->query($H);if(!is_object($I))return
false;$K=$I->fetch_row();return($K?$K[$l]:false);}function
get_vals($H,$c=0){$J=array();$I=connection()->query($H);if(is_object($I)){while($K=$I->fetch_row())$J[]=$K[$c];}return$J;}function
get_key_vals($H,$f=null,$ii=true){$f=connection($f);$J=array();$I=$f->query($H);if(is_object($I)){while($K=$I->fetch_row()){if($ii)$J[$K[0]]=$K[1];else$J[]=$K[0];}}return$J;}function
get_rows($H,$f=null,$k="<p class='error'>"){$pb=connection($f);$J=array();$I=$pb->query($H);if(is_object($I)){while($K=$I->fetch_assoc())$J[]=$K;}elseif(!$I&&!$f&&$k&&(defined('Adminer\PAGE_HEADER')||$k=="-- "))echo$k.error()."\n";return$J;}function
unique_array($K,array$w){foreach($w
as$v){if(preg_match("~^(PRIMARY|UNIQUE)$~",$v["type"])&&!$v["partial"]){$J=array();foreach($v["columns"]as$x){if(!isset($K[$x]))continue
2;$J[$x]=$K[$x];}return$J;}}}function
escape_key($x){if(preg_match('(^([\w(]+)('.str_replace("_",".*",preg_quote(idf_escape("_"))).')([ \w)]+)$)',$x,$B))return$B[1].idf_escape(idf_unescape($B[2])).$B[3];return
idf_escape($x);}function
where(array$Z,array$m=array()){$J=array();foreach((array)$Z["where"]as$x=>$X){$x=bracket_escape($x,true);$c=escape_key($x);$l=idx($m,$x,array());$ad=$l["type"];$ye=$l&&(is_blob($l)||preg_match('~binary~',$ad));$J[]=$c.($ye&&!is_utf8($X)?" = ".driver()->quoteBinary($X):(JUSH=="sql"&&$ad=="json"?" = CAST(".q($X)." AS JSON)":(JUSH=="pgsql"&&preg_match('~^jsonb?$~',$l["full_type"])?"::jsonb = ".q($X)."::jsonb":(JUSH=="sql"&&is_numeric($X)&&preg_match('~\.~',$X)?" LIKE ".q($X):(JUSH=="mssql"&&strpos($ad,"datetime")===false?" LIKE ".q(preg_replace('~[_%[]~','[\0]',$X)):" = ".unconvert_field($l,q($X)))))));if(JUSH=="sql"&&preg_match('~char|text~',$ad)&&preg_match("~[^ -@]~",$X))$J[]="$c = ".q($X)." COLLATE ".charset(connection())."_bin";}foreach((array)$Z["null"]as$x)$J[]=escape_key($x)." IS NULL";return
implode(" AND ",$J);}function
where_columns(array$m){$J=array();foreach((array)$_GET["null"]as$x)$J[$x]=true;foreach((array)$_GET["where"]as$x=>$X){$x=bracket_escape($x,true);foreach($m
as$D=>$l){if($x==$D||strpos($x,idf_escape($D))!==false)$J[$D]=true;}}return$J;}function
where_check($X,array$m=array()){parse_str($X,$Ua);remove_slashes(array(&$Ua));return
where($Ua,$m);}function
where_link($s,$c,$Y,$og="="){$lg=($Y!==null?$og:"IS NULL");return"&where[$s][col]=".url_escape($c).($lg!=first(adminer()->operators())?"&where[$s][op]=".url_escape($lg):"")."&where[$s][val]=".url_escape($Y);}function
convert_fields(array$d,array$m,array$M=array()){$J="";foreach($d
as$x=>$X){if($M&&!in_array(idf_escape($x),$M))continue;$va=convert_field($m[$x]);if($va)$J
.=", $va AS ".idf_escape($x);}return$J;}function
cookie_path(){return
strtr(preg_replace('~\?.*~','',$_SERVER["REQUEST_URI"]),array(";"=>"%3B",","=>"%2C"));}function
cookie($D,$Y,$Ye=2592000){header("Set-Cookie: $D=".rawurlencode($Y).($Ye?"; expires=".gmdate("D, d M Y H:i:s",time()+$Ye)." GMT":"")."; path=".cookie_path().(HTTPS?"; secure":"").($D=="adminer_import"?"":"; HttpOnly")."; SameSite=lax",false);}function
get_url($Ij,$tb){$http_response_header=null;$Ec=array();set_error_handler(function($Dc,$k)use(&$Ec){$Ec[]=preg_replace('~^file_get_contents\([^)]*\):\s*~','',$k);return
true;});$J=file_get_contents($Ij,false,$tb);restore_error_handler();$Nd=(function_exists('http_get_last_response_headers')?http_get_last_response_headers():$http_response_header);return
array($J,(preg_match('~^HTTP/[\d.]+ (\d+)~',idx($Nd,0,''),$B)?$B[1]:''),(array)$Nd,($J===false?implode("\n",$Ec):''),);}function
get_settings($wb){parse_str($_COOKIE[$wb],$ji);return$ji;}function
get_setting($x,$wb="adminer_settings",$j=null){return
idx(get_settings($wb),$x,$j);}function
save_settings(array$ji,$wb="adminer_settings"){$Y=http_build_query($ji+get_settings($wb));cookie($wb,$Y);$_COOKIE[$wb]=$Y;}function
restart_session(){if(!ini_bool("session.use_cookies")&&(!function_exists('session_status')||session_status()==PHP_SESSION_NONE))session_start();}function
stop_session($md=false){$Lj=ini_bool("session.use_cookies");if(!$Lj||$md){session_write_close();if($Lj&&ini_set("session.use_cookies",'0')===false)session_start();}}function&get_session($x){return$_SESSION[$x][DRIVER][SERVER][$_GET["username"]];}function
set_session($x,$X){$_SESSION[$x][DRIVER][SERVER][$_GET["username"]]=$X;}function
auth_url($Vj,$N,$V,$i=null){$Hj=remove_from_uri(implode("|",array_keys(SqlDriver::$drivers))."|username|ext|".($i!==null?"db|":"").($Vj=='mssql'||$Vj=='pgsql'?"":"ns|").session_name());preg_match('~([^?]*)\??(.*)~',$Hj,$B);return"$B[1]?".(sid()?SID."&":"").($_GET["ext"]?"ext=".url_escape($_GET["ext"])."&":"").($Vj!="server"||$N!=""?url_escape($Vj)."=".url_escape($N)."&":"")."username=".url_escape($V).($i!=""?"&db=".url_escape($i):"").($B[2]?"&$B[2]":"");}function
is_ajax(){return($_SERVER["HTTP_X_REQUESTED_WITH"]=="XMLHttpRequest");}function
redirect($A,$C=null){if($C!==null){restart_session();$_SESSION["messages"][preg_replace('~^[^?]*~','',($A!==null?$A:$_SERVER["REQUEST_URI"]))][]=$C;}if($A!==null){if($A=="")$A=".";header("Location: $A");exit;}}function
query_redirect($H,$A,$C,$zh=true,$Lc=true,$Vc=false,$Zi=""){if($Lc){$zi=microtime(true);$Vc=!connection()->query($H);$Zi=format_time($zi);}$ti=($H?adminer()->messageQuery($H,$Zi,$Vc):"");if($Vc){adminer()->error
.=error().$ti.script("messagesPrint();")."<br>";return
false;}if($zh)redirect($A,$C.$ti);return
true;}class
Queries{static$queries=array();static$start=0;}function
queries($H){if(!Queries::$start)Queries::$start=microtime(true);Queries::$queries[]=(driver()->delimiter!=';'?$H:(preg_match('~;$~',$H)?"DELIMITER ;;\n$H;\nDELIMITER ":$H).";");return
connection()->query($H);}function
apply_queries($H,array$T,$Gc='Adminer\table'){foreach($T
as$R){if(!queries("$H ".$Gc($R)))return
false;}return
true;}function
queries_redirect($A,$C,$zh){$uh=implode("\n",Queries::$queries);$Zi=format_time(Queries::$start);return
query_redirect($uh,$A,$C,$zh,false,!$zh,$Zi);}function
format_time($zi){return
sprintf('%.3f s',max(0,microtime(true)-$zi));}function
relative_uri(){return
preg_replace_callback('~^[^?]*~',function($B){return
str_replace(":","%3A",$B[0]);},preg_replace('~^[^?]*/([^?]*)~','\1',$_SERVER["REQUEST_URI"]));}function
remove_from_uri($Ig=""){return
substr(preg_replace("~(?<=[?&])($Ig".(SID?"":"|".session_name()).")=[^&]*&~",'',relative_uri()."&"),0,-1);}function
get_files($x,$Ob=false){$cd=$_FILES[$x];if(!$cd)return
null;foreach($cd
as$x=>$X)$cd[$x]=(array)$X;$J=array();foreach($cd["error"]as$x=>$k){if($k)return$k;$D=$cd["name"][$x];$hj=$cd["tmp_name"][$x];$rb=file_get_contents($Ob&&preg_match('~\.gz$~',$D)?"compress.zlib://$hj":$hj);if($Ob){$zi=substr($rb,0,3);if(function_exists("iconv")&&preg_match("~^\xFE\xFF|^\xFF\xFE~",$zi))$rb=iconv("utf-16","utf-8",$rb);elseif($zi=="\xEF\xBB\xBF")$rb=substr($rb,3);}$J[]=array($D,$rb);}return$J;}function
get_file($x,$Ob=false,$Ub=""){$fd=get_files($x,$Ob);if(!is_array($fd))return$fd;$J='';foreach($fd
as$cd){$rb=$cd[1];$J
.=$rb;if($Ub)$J
.=(preg_match("($Ub\\s*\$)",$rb)?"":$Ub)."\n\n";}return$J;}function
upload_error($k){$sf=($k==UPLOAD_ERR_INI_SIZE?ini_get("upload_max_filesize"):0);return($k?'Unable to upload a file.'.($sf?" ".sprintf('Maximum allowed file size is %sB.',$sf):""):'File does not exist.');}function
repeat_pattern($Wg,$y){return
str_repeat("$Wg{0,65535}",$y/65535)."$Wg{0,".($y%65535)."}";}function
is_utf8($X){return(preg_match('~~u',$X)&&!preg_match('~[\0-\x8\xB\xC\xE-\x1F]~',$X));}function
format_number($X){return
strtr(number_format($X,0,".",','),preg_split('~~u','0123456789',-1,PREG_SPLIT_NO_EMPTY));}function
format_status(array$S,$x){$X=idx($S,$x,'?');if(!is_numeric($X))return
h($X);if($X<0)return'?';$ra=($x=="Rows"&&(JUSH=="sqlite"||$S["Engine"]==(JUSH=="pgsql"?"table":"InnoDB")));return($ra?"~ ":"").format_number($X);}function
friendly_url($X){return
preg_replace('~\W~i','-',$X);}function
table_status1($R,$Wc=false){$J=table_status($R,$Wc);return($J?reset($J):array("Name"=>$R));}function
column_foreign_keys($R){$J=array();foreach(adminer()->foreignKeys($R)as$o){foreach($o["source"]as$X)$J[$X][]=$o;}return$J;}function
fields_from_edit(){$J=array();foreach((array)$_POST["field_keys"]as$x=>$X){if($X!=""){$X=bracket_escape($X);$_POST["function"][$X]=$_POST["field_funs"][$x];$_POST["fields"][$X]=$_POST["field_vals"][$x];}}foreach((array)$_POST["fields"]as$x=>$X){$D=bracket_escape($x,true);$J[$D]=array("field"=>$D,"full_type"=>"","type"=>"","privileges"=>array("insert"=>1,"update"=>1,"where"=>1,"order"=>1),"null"=>true,"auto_increment"=>($D==driver()->primary),);}return$J;}function
dump_headers($Yd,$Kf=false){$J=adminer()->dumpHeaders($Yd,$Kf);$Fg=$_POST["output"];if($Fg!="text"||$J=="tar"){$mb=($Fg!="text"&&$Fg!="file"&&preg_match('~^[0-9a-z]+$~',$Fg)?".$Fg":"");header("Content-Disposition: attachment; filename=".adminer()->dumpFilename($Yd).".$J$mb");}session_write_close();if(!ob_get_level())ob_start(null,4096);ob_flush();flush();return$J;}function
dump_csv(array$K){$uj=$_POST["format"]=="tsv";foreach($K
as$x=>$X){if(preg_match('~["\n]|^0[^.]|\.\d*0$|'.($uj?'\t':'[,;]|^$').'~',$X))$K[$x]='"'.str_replace('"','""',$X).'"';}echo
implode(($_POST["format"]=="csv"?",":($uj?"\t":";")),$K)."\r\n";}function
parse_csv($Db,$ei){$J=array();preg_match_all('~(?>"[^"]*"|[^"\r\n]+)+~',$Db,$if);foreach($if[0]as$K){preg_match_all("~((?>\"[^\"]*\")+|[^$ei]*)$ei~",$K.$ei,$jf);$J[]=$jf[1];}return$J;}function
csv_value($X){return(preg_match('~^".*"$~s',$X)?str_replace('""','"',substr($X,1,-1)):$X);}function
apply_sql_function($q,$c){return($q?($q=="unixepoch"?"DATETIME($c, '$q')":($q=="count distinct"?"COUNT(DISTINCT ":strtoupper("$q("))."$c)"):$c);}function
get_temp_dir(){return
ini_get("upload_tmp_dir")?:sys_get_temp_dir();}function
file_open_lock($n){if(is_link($n))return;$p=@fopen($n,"c+");if(!$p)return;@chmod($n,0660);if(!flock($p,LOCK_EX)){fclose($p);return;}return$p;}function
file_write_unlock($p,$Hb){rewind($p);fwrite($p,$Hb);ftruncate($p,strlen($Hb));file_unlock($p);}function
file_unlock($p){flock($p,LOCK_UN);fclose($p);}function
first(array$ua){return
reset($ua);}function
password_file($g){$n=get_temp_dir()."/adminer.key";if(!$g&&!file_exists($n))return'';$p=file_open_lock($n);if(!$p)return'';$J=stream_get_contents($p);if(!$J){$J=rand_string();file_write_unlock($p,$J);}else
file_unlock($p);return$J;}function
rand_string(){return(function_exists('random_bytes')?bin2hex(random_bytes(16)):md5(uniqid(strval(mt_rand()),true)));}function
select_value($X,$_,array$l,$Yi){if(is_array($X)){$J="";if(array_filter($X,'is_array')==array_values($X)){$Ie=array();foreach($X
as$W)$Ie+=array_fill_keys(array_keys($W),null);foreach(array_keys($Ie)as$He)$J
.="<th>".h($He);foreach($X
as$W){$J
.="<tr>";foreach(array_merge($Ie,$W)as$Qj)$J
.="<td>".select_value($Qj,$_,$l,$Yi);}}else{foreach($X
as$He=>$W)$J
.="<tr>".($X!=array_values($X)?"<th>".h($He):"")."<td>".select_value($W,$_,$l,$Yi);}return"<table>$J</table>";}if(!$_)$_=adminer()->selectLink($X,$l);if($_===null){if(is_mail($X))$_="mailto:$X";if(is_url($X))$_=$X;}$X=driver()->value($X,$l);$J=adminer()->editVal($X,$l);if($J!==null){if(!is_utf8($J))$J="\0";elseif($Yi!=""&&is_shortable($l))$J=shorten_utf8($J,max(0,+$Yi));else$J=h($J);}return
adminer()->selectVal($J,$_,$l,$X);}function
is_blob(array$l){return
preg_match('~blob|bytea|raw|file'.(JUSH=="mssql"?'|binary|image':'').'~',$l["type"])&&!in_array($l["type"],idx(driver()->structuredTypes(),'User types',array()));}function
is_mail($vc){$xa='[-a-z0-9!#$%&\'*+/=?^_`{|}~]';$kc='[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])';$Wg="$xa+(\\.$xa+)*@($kc?\\.)+$kc";return
is_string($vc)&&preg_match("(^$Wg(,\\s*$Wg)*\$)i",$vc);}function
is_url($Q){$kc='[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])';return
preg_match("~^((https?):)?//($kc?\\.)+$kc(:\\d+)?(/.*)?(\\?.*)?(#.*)?\$~i",$Q);}function
is_shortable(array$l){return!preg_match('~'.number_type().'|date|time|year~',$l["type"]);}function
host_port($N){return(preg_match('~^(:([^:].*)|(\[(.+)\]|(([^:]+://)?[^:]+))(:(\d+))?)$~',$N,$B)?array($B[4].$B[5],$B[2].$B[8]):array($N,''));}function
count_rows($R,array$Z,$ze,array$r){$H=" FROM ".table($R).($Z?" WHERE ".implode(" AND ",$Z):"");return($ze&&(JUSH=="sql"||count($r)==1)?"SELECT COUNT(DISTINCT ".implode(", ",$r).")$H":"SELECT COUNT(*)".($ze?" FROM (SELECT 1$H GROUP BY ".implode(", ",$r).") x":$H));}function
slow_query($H){$i=adminer()->database();$aj=adminer()->queryTimeout();$ni=driver()->slowQuery($H,$aj);$f=null;if(!$ni&&support("kill")){$f=connect();if($f&&($i==""||$f->select_db($i))){$Je=get_val(connection_id(),0,$f);echo
script("const timeout = setTimeout(() => { ajax('".js_escape(ME)."script=kill', function () {}, 'kill=$Je&token=".get_token()."'); }, 1000 * $aj);");}}ob_flush();flush();$J=@get_key_vals(($ni?:$H),$f,false);if($f){echo
script("clearTimeout(timeout);");ob_flush();flush();}return$J;}function
get_token(){$xh=rand(1,1e6);return($xh^$_SESSION["token"]).":$xh";}function
verify_token(){list($ij,$xh)=explode(":",$_POST["token"]);return($xh^$_SESSION["token"])==$ij&&in_array($_SERVER["HTTP_SEC_FETCH_SITE"],array("","same-origin"));}function
compress_alphabet(){return
strtr(implode(range('"','~')),"'\\","!\n");}function
decompress_string($Q,$ac=""){$oa=array_flip(str_split(compress_alphabet()));$y=strlen($Q);$Sj=($y?13*($y-1)/2-$oa[$Q[0]]:0);$Ia="";$Kh=0;$Lh=0;for($s=1;$s<$y;$s+=2){$Kh=($Kh<<13)+$oa[$Q[$s]]*93+$oa[$Q[$s+1]];$Lh+=13;while($Lh>=8&&$Sj>=8){$Lh-=8;$Sj-=8;$Ia
.=chr($Kh>>$Lh);$Kh&=(1<<$Lh)-1;}}if($Ia=="")return"";if($ac!=""&&function_exists('inflate_init'))return
inflate_add(inflate_init(ZLIB_ENCODING_RAW,array('dictionary'=>$ac)),$Ia,ZLIB_FINISH);return($ac==""&&function_exists('gzinflate')?gzinflate($Ia):inflate($Ia,$ac));}function
inflate($Ia,$ac=""){$Ve=array(3,4,5,6,7,8,9,10,11,13,15,17,19,23,27,31,35,43,51,59,67,83,99,115,131,163,195,227,258);$We=array(0,0,0,0,0,0,0,0,1,1,1,1,2,2,2,2,3,3,3,3,4,4,4,4,5,5,5,5,0);$ec=array(1,2,3,4,5,7,9,13,17,25,33,49,65,97,129,193,257,385,513,769,1025,1537,2049,3073,4097,6145,8193,12289,16385,24577);$gc=array(0,0,0,0,1,1,2,2,3,3,4,4,5,5,6,6,7,7,8,8,9,9,10,10,11,11,12,12,13,13);$J=$ac;$G=0;do{$hd=inflate_bits($Ia,$G,1);$U=inflate_bits($Ia,$G,2);if(!$U){$G=($G+7)&~7;$y=inflate_bits($Ia,$G,16);$G+=16;$J
.=substr($Ia,$G>>3,$y);$G+=$y<<3;}else{if($U==1){$cf=array_merge(array_fill(0,144,8),array_fill(0,112,9),array_fill(0,24,7),array_fill(0,8,8));$hc=array_fill(0,30,5);}else{$bf=inflate_bits($Ia,$G,5)+257;$fc=inflate_bits($Ia,$G,5)+1;$E=array(16,17,18,0,8,7,9,6,10,5,11,4,12,3,13,2,14,1,15);$Df=array_fill(0,19,0);$Cf=inflate_bits($Ia,$G,4)+4;for($s=0;$s<$Cf;$s++)$Df[$E[$s]]=inflate_bits($Ia,$G,3);$Ef=inflate_table($Df);$Xe=array();while(count($Xe)<$bf+$fc){$Ii=inflate_symbol($Ia,$G,$Ef);if($Ii==16)$Xe=array_merge($Xe,array_fill(0,inflate_bits($Ia,$G,2)+3,end($Xe)));elseif($Ii==17)$Xe=array_merge($Xe,array_fill(0,inflate_bits($Ia,$G,3)+3,0));elseif($Ii==18)$Xe=array_merge($Xe,array_fill(0,inflate_bits($Ia,$G,7)+11,0));else$Xe[]=$Ii;}$cf=array_slice($Xe,0,$bf);$hc=array_slice($Xe,$bf);}$df=inflate_table($cf);$jc=inflate_table($hc);while(($Ii=inflate_symbol($Ia,$G,$df))!=256){if($Ii<256)$J
.=chr($Ii);else{$y=$Ve[$Ii-257]+inflate_bits($Ia,$G,$We[$Ii-257]);$ic=inflate_symbol($Ia,$G,$jc);$dg=strlen($J)-$ec[$ic]-inflate_bits($Ia,$G,$gc[$ic]);for($s=0;$s<$y;$s++)$J
.=$J[$dg+$s];}}}}while(!$hd);return($ac==""?$J:substr($J,strlen($ac)));}function
inflate_bits($Ia,&$G,$yb){$J=0;for($s=0;$s<$yb;$s++){$J+=((ord($Ia[$G>>3])>>($G&7))&1)<<$s;$G++;}return$J;}function
inflate_table(array$Xe){$R=array();$cb=0;for($Ja=1;$Ja<=max($Xe);$Ja++){foreach($Xe
as$Ii=>$y){if($y==$Ja){$R[$Ja][$cb]=$Ii;$cb++;}}$cb<<=1;}return$R;}function
inflate_symbol($Ia,&$G,array$R){$cb=0;$Ja=0;do{$cb=($cb<<1)+inflate_bits($Ia,$G,1);$Ja++;}while(!isset($R[$Ja][$cb]));return$R[$Ja][$cb];}function
script($ri,$lj="\n"){return"<script".nonce().">$ri</script>$lj";}function
script_src($Ij,$Rb=false){return"<script src='".h($Ij)."'".nonce().($Rb?" defer":"")."></script>\n";}function
nonce(){return' nonce="'.get_nonce().'"';}function
on($Hc,$Gd,$sa=null){$ta=array();foreach(array_slice(func_get_args(),2)as$X)$ta[]=json_encode($X,256);return" data-on$Hc='".str_replace(array('&','<',"'"),array('&amp;','&lt;','&#039;'),"$Gd(".implode(", ",$ta).")")."'";}function
input_hidden($D,$Y=""){return"<input type='hidden' name='".h($D)."' value='".h($Y)."'>\n";}function
input_token(){return
input_hidden("token",get_token());}function
target_blank(){return' target="_blank" rel="noreferrer noopener"';}function
h($Q){return
str_replace(array('&','<','"',"'","\0"),array('&amp;','&lt;','&quot;','&#039;','&#0;'),$Q);}function
nl_br($Q){return
str_replace("\n","<br>",$Q);}function
checkbox($D,$Y,$Wa,$Le="",$b="",$bb="",$Ne=""){$J="<input type='checkbox' name='$D' value='".h($Y)."'".($Wa?" checked":"").($Le==""&&$bb?" class='$bb'":"").($Ne?" aria-labelledby='$Ne'":"").$b.">";return($Le!=""?"<label".($bb?" class='$bb'":"").">$J".h($Le)."</label>":$J);}function
optionlist($sg,$bi=null,$Mj=false){$J="";foreach($sg
as$He=>$W){$tg=array($He=>$W);if(is_array($W)){$J
.='<optgroup label="'.h($He).'">';$tg=$W;}foreach($tg
as$x=>$X)$J
.='<option'.($Mj||is_string($x)?' value="'.h($x).'"':'').($bi!==null&&($Mj||is_string($x)?(string)$x:$X)===$bi?' selected':'').'>'.h($X);if(is_array($W))$J
.='</optgroup>';}return$J;}function
html_select($D,array$sg,$Y="",$b="",$Ne=""){static$Le=0;$Me="";if(!$Ne&&substr($sg[""],0,1)=="("){$Le++;$Ne="label-$Le";$Me="<option value='' id='$Ne'>".h($sg[""]);unset($sg[""]);}return"<select name='".h($D)."'".($Ne?" aria-labelledby='$Ne'":"")."$b>".$Me.optionlist($sg,$Y)."</select>";}function
html_radios($D,array$sg,$Y="",$ei=""){$J="";foreach($sg
as$x=>$X)$J
.="<label><input type='radio' name='".h($D)."' value='".h($x)."'".($x==$Y?" checked":"").">".h($X)."</label>$ei";return$J;}function
confirm($C=""){return
on('click','confirmClick',$C?:'Are you sure?');}function
print_fieldset($t,$Ue,$Zj=false){echo"<fieldset><legend>","<a href='#fieldset-$t' class='toggle'>$Ue</a>","</legend>","<div id='fieldset-$t'".($Zj?"":" class='hidden'").">\n";}function
bold($La,$bb=""){return($La?" class='active $bb'":($bb?" class='$bb'":""));}function
js_escape($Q){return
str_replace("<","\\x3C",addcslashes($Q,"\r\n'\\"));}function
js_escape_re($Q){return
addcslashes(preg_quote($Q,"/"),"\r\n");}function
pagination_href($F){return
remove_from_uri("page|next").($F?"&page=$F".($_GET["next"]!=""?"&next=".url_escape($_GET["next"]):""):"");}function
pagination($F,$Eb){return" ".($F==$Eb?($F?"<b>".($F+1)."</b>":$F+1):'<a href="'.h(pagination_href($F)).'">'.($F+1)."</a>");}function
hidden_fields(array$rh,array$be=array(),$kh=''){$J=false;foreach($rh
as$x=>$X){if(!in_array($x,$be)){if(is_array($X))hidden_fields($X,array(),$x);else{$J=true;echo
input_hidden(($kh?$kh."[$x]":$x),$X);}}}return$J;}function
hidden_fields_get(){echo(sid()?input_hidden(session_name(),session_id()):''),($_GET["ext"]?input_hidden("ext",$_GET["ext"]):""),(isset($_GET[DRIVER])?input_hidden(DRIVER,SERVER):""),input_hidden("username",$_GET["username"]);}function
file_input($b,$Kh=""){$mf="max_file_uploads";$nf=ini_get($mf);$sf="upload_max_filesize";$tf=ini_bytes($sf);$hh=ini_bytes("post_max_size");if($hh&&$hh<$tf){$sf="post_max_size";$tf=$hh;}$uf=ini_get($sf);return(ini_bool("file_uploads")?"<input type='file'$b".on('change','fileChange',(int)$nf,sprintf('Increase %s.',"$mf = $nf"),$tf,sprintf('Increase %s.',"$sf = $uf")).">$Kh":'File uploads are disabled.');}function
enum_input($U,$b,array$l,$Y,$yc=""){preg_match_all("~'((?:[^']|'')*)'~",$l["length"],$if);$kh=($l["type"]=="enum"?"val-":"");$Wa=(is_array($Y)?in_array("null",$Y):$Y===null);$J=($l["null"]&&$kh?"<label><input type='$U'$b value='null'".($Wa?" checked":"")."><i>$yc</i></label>":"");foreach($if[1]as$X){$X=stripcslashes(str_replace("''","'",$X));$Wa=(is_array($Y)?in_array($kh.$X,$Y):$Y===$X);$J
.=" <label><input type='$U'$b value='".h($kh.$X)."'".($Wa?' checked':'').'>'.h(adminer()->editVal($X,$l)).'</label>';}return$J;}function
input(array$l,$Y,$q,$Aa=false,$Fj=false){$D=h(bracket_escape($l["field"]));echo"<td class='function'>";if(is_array($Y)&&!$q)$q="json";$Ee=($q=="json"||preg_match('~^jsonb?$~',$l["full_type"]));if($Ee&&$Y!=''&&(JUSH!="pgsql"||$l["type"]!="json"))$Y=json_encode(is_array($Y)?$Y:json_decode($Y),128|64|256);$Jh=(JUSH=="mssql"&&$Fj&&$l["auto_increment"]);if($Jh&&!$_POST["save"])$q=null;$xd=(isset($_GET["select"])||$Jh?array("orig"=>'original'):array())+adminer()->editFunctions($l);$Cc=driver()->enumLength($l);if($Cc){$l["type"]="enum";$l["length"]=$Cc;}$b=" name='fields[$D]".($l["type"]=="enum"||$l["type"]=="set"?"[]":"")."'".($Aa?" autofocus":"");echo
driver()->unconvertFunction($l)." ";$R=$_GET["edit"]?:$_GET["select"];if($l["type"]=="enum")echo
h($xd[""])."<td>".adminer()->editInput($R,$l,$b,$Y);else{$Id=(in_array($q,$xd)||isset($xd[$q]));$id=0;foreach($xd
as$x=>$X){if($x===""||!$X)break;$id++;}echo(count($xd)>1?"<select name='function[$D]'".on('change','functionChange').on_help_value('^SQL$').">".optionlist($xd,$q===null||$Id?$q:"")."</select>":h(reset($xd)))."<td".($id&&count($xd)>1?on('input','skipOriginal',$id):"").">";$qe=adminer()->editInput($R,$l,$b,$Y);if($qe!="")echo$qe;elseif(preg_match('~bool~',$l["type"]))echo"<input type='hidden'$b value='0'>"."<input type='checkbox'".(preg_match('~^(1|t|true|y|yes|on)$~i',$Y)?" checked":"")."$b value='1'>";elseif($l["type"]=="set")echo
enum_input("checkbox",$b,$l,(is_string($Y)?explode(",",$Y):$Y));elseif(is_blob($l)&&ini_bool("file_uploads"))echo"<input type='file' name='fields-$D'>";elseif($Ee)echo"<textarea$b cols='50' rows='12' class='jush-json'>".h($Y).'</textarea>';elseif(($Xi=preg_match('~text|lob|memo~i',$l["type"]))||preg_match("~\n~",$Y)){if($Xi&&JUSH!="sqlite")$b
.=" cols='50' rows='12'";else{$L=min(12,substr_count($Y,"\n")+1);$b
.=" cols='30' rows='$L'";}echo"<textarea$b>".h($Y).'</textarea>';}else{$xj=driver()->types();$vf=(!preg_match('~int~',$l["type"])&&preg_match('~^(\d+)(,(\d+))?$~',$l["length"],$B)?((preg_match("~binary~",$l["type"])?2:1)*$B[1]+($B[3]?1:0)+($B[2]&&!$l["unsigned"]?1:0)):($xj[$l["type"]]?$xj[$l["type"]]+($l["unsigned"]?0:1):0));if(JUSH=='sql'&&min_version(5.6)&&preg_match('~time~',$l["type"]))$vf+=7;echo"<input".((!$Id||$q==="")&&preg_match('~^'.int_type().'$~',$l["type"])&&!preg_match('~\[]~',$l["full_type"])?" type='number'":"")." value='".h($Y)."'".($vf?" data-maxlength='$vf'":"").(preg_match('~char|binary~',$l["type"])&&$vf>20?" size='".($vf>99?60:40)."'":"")."$b>";}echo
adminer()->editHint($R,$l,$Y),(count($xd)>1?script("fire(qs('select', qsl('td').previousSibling), 'change');",""):"");}}function
process_input(array$l){$u=bracket_escape($l["field"]);$q=idx($_POST["function"],$u);if($q=="orig")return(preg_match('~^CURRENT_TIMESTAMP~i',$l["on_update"])?idf_escape($l["field"]):false);if($q=="NULL")return"NULL";if(is_blob($l)&&ini_bool("file_uploads")){$cd=get_file("fields-$u");if(!is_string($cd))return
false;return
driver()->quoteBinary($cd);}$Y=idx($_POST["fields"],$u);if($Y===null)return
false;if($l["type"]=="enum"||driver()->enumLength($l)){$Y=idx($Y,0);if($Y=="orig"||!$Y)return
false;if($Y=="null")return"NULL";$Y=substr($Y,4);}if($l["auto_increment"]&&$Y=="")return
null;if($l["type"]=="set")$Y=implode(",",(array)$Y);if($q=="json"){$Y=json_decode($Y,true);if(!is_array($Y))return
false;return$Y;}return
adminer()->processInput($l,$Y,$q);}function
search_tables(){$_GET["where"][0]["val"]=$_POST["query"];$di="<ul>\n";foreach(table_status('',true)as$R=>$S){$D=adminer()->tableName($S);if(isset($S["Engine"])&&$D!=""&&(!$_POST["tables"]||in_array($R,$_POST["tables"]))){$I=connection()->query("SELECT".limit("1 FROM ".table($R)," WHERE ".implode(" AND ",adminer()->selectSearchProcess(fields($R),array())),1));if(!$I||$I->fetch_row()){$nh="<a href='".h(ME."select=".url_escape($R)."&where[0][op]=".url_escape($_GET["where"][0]["op"])."&where[0][val]=".url_escape($_GET["where"][0]["val"]))."'>$D</a>";echo"$di<li>".($I?$nh:"<p class='error'>$nh: ".error())."\n";$di="";}}}echo($di?"<p class='message'>".'No tables.':"</ul>")."\n";}function
on_help($Xi,$li=0){return
on('mouseover','helpMouseover',$Xi,$li).on('mouseout','helpMouseout');}function
on_help_value($Fh="",$Ih=""){return
on('mouseover','helpValueMouseover',$Fh,$Ih).on('mouseout','helpMouseout');}function
edit_form($R,array$m,$K,$Fj,$k=''){$Li=adminer()->tableName(table_status1($R,true));page_header(($Fj?'Edit':'Insert'),$k,array("select"=>array($R,$Li)),$Li);adminer()->editRowPrint($R,$m,$K,$Fj);if($K===false){echo"<p class='error'>".'No rows.'."\n";return;}echo"<form action='' method='post' enctype='multipart/form-data' id='form'>\n";$tc=false;$fk=($Fj&&!isset($_GET["select"])?where_columns($m):array());$ub=(count($fk)!=count($m));if(!$ub)$fk=array();if(!$m)echo"<p class='error'>".'You have no privileges to update this table.'."\n";else{echo"<table class='layout nowrap'".on('keydown','editingKeydown').">\n";$Aa=!$_POST;foreach($m
as$D=>$l){echo"<tr".($fk[$D]?on('change','whereChange'):"")."><th>".adminer()->fieldName($l);$j=idx($_GET["set"],bracket_escape($D));if($j===null){$j=$l["default"];if($l["type"]=="bit"&&preg_match("~^b'([01]*)'\$~",$j,$Gh))$j=$Gh[1];if(JUSH=="sql"&&preg_match('~binary~',$l["type"]))$j=bin2hex($j);}$Y=($K!==null?($K[$D]!=""&&JUSH=="sql"&&preg_match("~enum|set~",$l["type"])&&is_array($K[$D])?implode(",",$K[$D]):(is_bool($K[$D])?+$K[$D]:$K[$D])):(!$Fj&&$l["auto_increment"]?"":(isset($_GET["select"])?false:$j)));if(!$_POST["save"]&&is_string($Y))$Y=adminer()->editVal($Y,$l);if(($Fj&&!isset($l["privileges"]["update"]))||$l["generated"])echo"<td class='function'><td>".select_value($Y,'',$l,null);else{$tc=true;$q=($_POST["save"]?idx($_POST["function"],bracket_escape($D),""):($Fj&&preg_match('~^CURRENT_TIMESTAMP~i',$l["on_update"])?"now":($Y===false?null:($Y!==null?'':'NULL'))));if(!$_POST&&!$Fj&&$Y==$l["default"]&&preg_match('~^[\w.]+\(~',$Y))$q="SQL";if(preg_match("~time~",$l["type"])&&preg_match('~^CURRENT_TIMESTAMP~i',$Y)){$Y="";$q="now";}if($l["type"]=="uuid"&&$Y=="uuid()"){$Y="";$q="uuid";}if($Aa!==false)$Aa=($l["auto_increment"]||$q=="now"||$q=="uuid"?null:true);input($l,$Y,$q,$Aa,$Fj);if($Aa)$Aa=false;}}if(!fields($R)&&driver()->primary!="")echo"<tr>"."<th><input name='field_keys[]'".on('input','fieldChange').">"."<td class='function'>".html_select("field_funs[]",adminer()->editFunctions(array("null"=>isset($_GET["select"]))))."<td><input name='field_vals[]'>";echo"</table>\n";}echo"<p>\n";if($tc){echo"<input type='submit' value='".'Save'."'>\n";if(!isset($_GET["select"])&&$ub){$bc=($fk&&($k!=""||adminer()->error!="")?" disabled":"");echo"<input type='submit' name='insert' value='".($Fj?'Save and continue edit':'Save and insert next')."' title='Ctrl+Shift+Enter'$bc".($Fj?on('click','ajaxForm','Saving…'):"").">\n";}}echo($Fj?"<input type='submit' name='delete' value='".'Delete'."'".confirm().">\n":"");if(isset($_GET["select"]))hidden_fields(array("check"=>(array)$_POST["check"],"clone"=>$_POST["clone"],"all"=>$_POST["all"]));echo
input_hidden("referer",(isset($_POST["referer"])?$_POST["referer"]:$_SERVER["HTTP_REFERER"])),input_hidden("save",1),input_token(),"</form>\n";}function
shorten_utf8($Q,$y=80,$Ei=""){if(!preg_match("(^(".repeat_pattern("[\t\r\n -\x{10FFFF}]",$y).")($)?)u",$Q,$B))preg_match("(^(".repeat_pattern("[\t\r\n -~]",$y).")($)?)",$Q,$B);return
h($B[1]).$Ei.(isset($B[2])?"":"<i>…</i>");}function
icon($Xd,$D,$Wd,$cj,$b=""){return"<button ".($D?"type='submit' name='$D'":"draggable='true' tabindex='-1'")." title='".h($cj)."' class='icon icon-$Xd".($D?"":" jsonly")."'$b><span>$Wd</span></button>";}function
copy_icon(){$xb='Copy';return"<a href='' class='jsonly icon-copy' title='$xb'><span>$xb</span></a>";}if(isset($_GET["file"])){if(substr(VERSION,-4)!='-dev'){if($_SERVER["HTTP_IF_MODIFIED_SINCE"]){header("HTTP/1.1 304 Not Modified");exit;}header("Expires: ".gmdate("D, d M Y H:i:s",time()+365*24*60*60)." GMT");header("Last-Modified: ".gmdate("D, d M Y H:i:s")." GMT");header("Cache-Control: immutable");}ini_set("zlib.output_compression",'1');if($_GET["file"]=="default.css"){header("Content-Type: text/css; charset=utf-8");echo
decompress_string('!c0=@iDZ*tV?H*{U)[Q;B/1SR=Dh9&hJv;rrHHN,.V&KGmzhDwb9E:tfItN#CwUSwX?Xyeqi5d/N>]A"1lTaK
Tx^G#)>.UM~&(MUO{shFwKG+g4,>C*S:
f1hRcL)KhkmZFtH^qWCMBf7tZ{.#f{8V6<
#Nk9.jSA&0km
lxTc6$tVXF.+.*cJeW<wG~51NPIP4xT,`5Fw(3!{(~-,9<s}YqWT+L%^[i[s<&8ErH[O8<a)
ljb
$LurL4t]W%a>H/b
X/{EMCz:LXX((.yD>6A0+]t%ACU_
:"Bp%c=`r4T.#6G1(p
xo=TMNIiX,W0G-OEkD}^/L"3iRuM0)KZQ^aWB9dsO%0WmcO<LgliJIDSwKw0uo4(Piokl7g)}Qq_R"C>
^,?D183n.@41e}1M3L@&rCG$;yG3^fAu1qCeb_`V5R)ywQ+^^}Y?,S-#YZFZG
*@I%I_vxm,Tu:<aGT4wdZr#t8h]Nq~_-mA_aP)C2W
3#$o
g`gA/T"apmp;"31><i"",jWq9Wx4|Kj$:Svf`fH`|l`L/=wn!GzOm+(2zYb@S?I6~Dgg51]s$GQ<f%*sZ)4os*u%H<]daIUU7+nOS>!R,?jI-ZyOTT8YA+<ro/FX
5%v%]D1&UG`Rk{"Wc.*PH+X"!Vb@SA=T#6)+N_
VgZ:[vm-?:-d-#LVMbB`M*
o3=!8PG}PV45(W`#.!4Aj#=
`|=e];={gdf>3&l{-kM.$C*+s{3":?S*Zv4|Rl!*UYvBXH@}(A,#om09^h1i;#LHmj2,KUT]s;#mi|*91KjF]nE
u?>^sG`oFK
)Wofomi0<!n"hdYaSs6[44(o8rHBG_@1V2u@D_*jz/#ZgKg<,ob6)a>B~0
Nc9PJ]bx=7K{0`!<w~"{8gg&A2+L#$C,xw#&#5qLhH:Y
6oD1wS)Hu:z&]%$L:*RH&&hm9*p.)J&x-8E0z+soB4Y.o:5!`DtOyw7783CWgj
WZ%`ELhCb!<9!`!t;k@5]}^L$~J|@.agt}?B1>
;"ZGyF-kRn"BIwMi;iFn0;f?!>s@V%wLZZ[kKwyDKfGko5=+|UjHeZjXy;0;#G@L"d`Um3u4Z@)WU.Kf:>6w?u|8l*.uRy`amgR$8Nv?MAbetW1fZC=.a/i!<lm+CgiuJdI)Ig2l@6xS*[@!@B.hXtesj)KZ`"D(QZyUo#,afykRAvt+#nz?,6c9u&`9kdt)X35?[Y<n!"C4r!0$AJl3>+#H(mk1pQn,Z3ZZ]8D)q@wst)_4|I5f}dsW#hqo*!.
4#"_/$:mCAq.5UCVL;oIlfE&U`w!f9m@e?)4t)~-8Kr.@Bm$9-|,R
ult.W=H4(dSM?+2D
gapxO[e_/=:kYP05Q|i+[N_N-YHIAe)0I*>{8]&Z?!aMCh.,oL@6p&lY$/
U=Fds9>*<FAc!>,5<A*;C+!_3@O6|?
//8+>*;@Um0hT[y8<Yt,@dvwiU()maH>967;d_]`={>5EWy&s32*#uINV)k5YG"ekF2}hI1O:Mj?8&AG/j[.-n!P5/("uWRm`3"j5%iI+qc5SJ+:9eOv83%i]U%[V*dHY/2lm8EP@h:*ITM4#//"X9KhV|EJ;q*De_$X_uTkg^D"0(-oU$AjZ^;N#1fw]3U1a`)mkv^ymmdQDDS;q71|/~(_`/BNq++E<jkdNVV@mh?.W_4M<w=_(ybU80Bn/@V^N!54"@
H!U(`":dm[rVdms6(S],j-batnN&O(^ru_<To+HJ~-NHu&
@=6dl$NMZ6.-yJ1jkQMe$lOAr{RGpt_56jj]YcYfGIIoQ"Gp(LKTcE%
(#A-Ss>OBN92I3S^/SZ[Y{fR5U]f3~"At`%-@82:PR%ue
?eN{]g1_DyuG)bqfX_Qs^{78KXKDK$X3a[Ecg5g;AJ4X-!D6[i7{=;"[e<PUxWs#8`
A
"RPK}9NBKH69/U1HuFO6us<2>Oj!}^0)84[gIeYd+TUi:!R?Fdgb1)D%@K@H-J^5m+l!YgY?3xK#mU+#Qm]0g"V)XSM0|uEUQFXQBk%K.*BURB
10iC:gT=p:fn<{1$Zp5Ovs@e;!HiB|gIW:-0v}QHRYkl1<T*o2TX81n2`o1uT)@|7jq%"/7@G)frLDr8H3rf_Na<86aPUN*MUbTee4EdGd"1t=NaEF
;iySU?Y;v.4>
RP2b=1]Ad#4rICDk=x"Z`z?~X!`9<#2>$)ds-gMK0Ux00NP!8ZqDXXp3GXgn;Q/Yh-#qdOQ)S}u]q?W-CFEV(7eTC:uKZPMYS5$i,I
2[0ov7,YH0pH6R`n,nr(8U[C-nTMc4~8NY~#n])3`!4/vklG
N%[#+;)i29O[fww}VbQ]rKi%y-ZW>Gfjs~p(*;U,"!Nb?-U]j|.+n]tcTM&fIT9Txd(Xod^%"{+[N9i2wzya_4MaMv!mcFuYxzK2uWypf-Yk^aYCxviP2qUT6}5(x~cMiq^HyEAUC"wnB}[#@KK)25n!nubey@KY7vpFug_hT>MRR-PyeDx`nbnx+Esrx8K7s3mrcwqM21p|blg?r41.s2Fe>OEZ`k/JpHne&Us&nDhatAWZv[y&$@`;Qn@WmZnmpc6]Y4TVtMp;4DQvyF0k8pZD=5bWwf
&@wXiEVG3L~r9x>f
DJv/ymXPyEB_ctnuw=s~Cw_nhh`{Ej14p4<A."no_E=r?V>
X/mcPhtkK
SHlA[=3)jxBjTN?Fe=K3mXc[JfhI)O+H?s4z>nCQ1zsK7lTpOp/EyNykw~PCe>j:YBo!af6g2s.elsKdH90yhlW|`Ib;pXh[h8+thI)*3^^VMelUX,D7=E%gNYB|!Ui(a%<e,qo(V~,OVWBQiQa{E~JjL.l)Z)s~SM<;1@H4S=U(RF;7uOW[""8dG;qK-cOo"FZYYl"dR%j%1)P33!1_jb2My9WJ$HcX;,>3z#p{Vl.-7WAo(snORM+[dimjqqHs6f=}g$mcji$bJZ[:XSHOb>Xmtx/<NJHc]$bOxnwttNq~T+A~i&.dMyF@w]c8_/6O8XCKwJyFbyBn[4e~Mb[,2oL?
nw.DpU|W%F+RddY/Zd.3$0W!sX9Lx^%b7@mO2x}SPPcu5gdBS;rHC:7^tfOQRKVQD&s)9oM`LZ_q;f]]#43gQkr=Z_L9p.?`P=j.yac@GbLQ^f<GGZ~]6kya;<8:F@hcNG,P!41@F)/:1Qq@&!L?]UN!3MG]s[~OAU`#nEbRUKQ*1(wpi7z-+cg_!Rbg$@^7a,TC#U_Is_vM"7L=&1(@d.AN*2H2hG(a<4JcT4QrvvmyR9>VCK0L;+;0K/BTI7@[lNzAgjG:)Y5`IIGxr!$lWV1aWii,3F!/{`s13+P
WmZ/+Rx>/5kfH^HZw96+`+Kg
d{skqlg_Z&_>xhPZN4XZ]LMvK_DQt{mZ>{EvrNTg9Ym"-?2=4,a]XVx$<gc$#&rc`0c@w"b
iJmT:GD~_
_*u!sl`,twN*82)AXsWx_7Chyn=14#77aNK<%RuEZMtV5bYfhw2RC,-2"(L,y<0T4qAY:ww{Z~<,Y>D}(Bk4n+&c"gcS0j9VkYQ=uJ<^:uQ
et&&B>f3yfuUqMh2DNQD;mk?ce`(?3S-WN4Q1mQBHMO(Gt2+nYFK`tncD)YIlKodjnrl]=O|6G5>V:Ex=5w]75MD501&2,gCqF
k]B*OK&O4=sBXd&q8BwW5hdkg4S1j#0X,E7vh8)NNmp6R]1curHvrKxu59J$S:1L6A;Mx/840/;]@BF)AkbJ~D"3>/"SoFH+&mO[y"EbOnY$evn)/<+-->rA,2QHBqRdJk;W4j#^S7AHal}2Hr|mYOBF`,Xw!J(7+RrSuO[aRC^yfg`Y^qE:o;0d/Y9/CaT?ZW{B^F/S1p:<?qS(|,a_M$6L5UzkJJXb=!>?qKg`yn*Wr@Nf1l/UF^3^*u?/3yJi?`lUmesKe.$4kv5*Ca(`?m[EX0/aO>(xbesGcRlE_Tzf7d<XX`)9G1`d>6VfK@Id
n;HULv1;x]a+n|H5$GpIYWE/2@2bH*b~Euozf/VRryH=4fpDuUa&LMWGpnMtJfxRu!YDf^_3r-=3nMt_v[3Z9s%iTc/<AA
"jF:YJ4c},/bNbT^[Ekpl"UA;J%)h%kQ4b*56aD9FCwnmmTb59/A%AR`60iM=y%Qxs_XZ)XQQ8O8pc7Z4-$?i5EYH$c#?7_-Yslj.FL$Er7
c$p(GPo*@RfuCsbme7[v:$ZGKTl"_M_)Ym0h+QDWXq?4Eoh-R7bn~ga%7@ZBwnyMh7a');}elseif($_GET["file"]=="dark.css"){header("Content-Type: text/css; charset=utf-8");echo
decompress_string('%OsbOb3V?!K0U*,j#-4V$+4lSl,oCh*02mX@fy~Y!-lFD?AZS5iE
nM`YKnnN5@7$,h]yHv0]"r/{_.;5=S+SNKE}<JYs`q%O%%)irj"Ua|G&>l)NqxPHIui")?!f$TF|nwt-nQCaG&Tzq)X$0a:"l<uhiWpN+Q>JUl.I??
0[m@%2{lZZ-SVaY0c(Abuipc;HrUB.?"L
&fe39+`O>CaP%DBGl_a;sKU:Vn{vUd)#z;(-4/lH:f/yqJRLo1D)]&Q)#F_Ex@I.Aoq!%P+x`#:u7a*NRit]e+S_#3_W;B1:p*qj1n&6tLeURFTa*Z%=PigZV?!E,M#fGWI7
Vby;v}uyiyNSk%!K32:q%~)Z7R]f7*[T1VD8GAHNE,gNAjPt3bJTq!),5tH82n<xEH5{06?o3=vyf/"d[Dx=^/`OW(R/VJpy<uN~pK
XY0h>3?PG;:6W2&H^g`XJac/.2vy_[sa[I@2XZ6h^)(qYAo-$5uc0Ep%,GX=n?^Dh<AHDPP6:^cBoLiHv;/&f"x+
Fxs2:m>cC)c>Lo
0T]2{suTY+[`^=g^8K@M"IJhD,eB]&O05-RUzKB;q=jP@t>t?wQJam-T
Ct4iGwsJeBb--L[GY@5KjZDe)KI2"iJ+I
sFktJV_tO_ae<,6L%wV]]G$83G65)NlCxcni0jK(!Hn+6;A/K;bfn9xSp=TVCsf``qH7Mimc,xAY3>O[u4w?fz&Fj
9f,[];aKusLC!8-;hiECD`(]x7[,6WvZQwb}-<xBhai*6z.x#y/,/,PfbzjJZY5<k)c%nD&#@k/fnmY$Bx2daWEELXWrfOnaM:!Fa_[qjXgtfwLcv6,3f~T:>3n3wR;MUKGkB;/1<=rsb.a0udo%x4L7HAUd8(4Q+6[S3/m5?BQNG}h9#D7rZ(C[A#`5XL+tAmR_k4;*wtK-0+ixONPclR9
9Q2)1cdU5,ODCdgYd!N6');}elseif($_GET["file"]=="functions.js"){header("Content-Type: text/javascript; charset=utf-8");echo
decompress_string('$c4]`nsWl1ptWOv_h:.%y>(B8Jhsf@ooc^S6O.cGC9BHMu=?3)
3[.X?=Wv,ZyTxSc%"*Tj_GrEE9;FU$:J.f=X0d>JYBIVZj]D<aA3aJq*XnQqxcw)y-@0VgkGu?I^TmUVgb:3EfFdr)MNb!A,_(k]_9iV_"_(G,09n1q_nV?v_Ya}yH922HAvE)Q-Dca{7S&0T[#Dtk/#Bl_W&wjz.9y.XC]Yq%Jhb/AiQuuuIe`
H[3+1Q4&&~EB[RStfa@oiB?KrvDpHs&j#r_LJL`Fqt*g"xA[A8
[y~qK:8it:6_]`i&KGku1;72%Z|S"<"lD2ac}UGi{G0+g.v*]W)w>x/UcKbq9g^FuU`Cx2;l0o]3G78`XOn?+m4h40J%^y[,j8/jf`[Aa
ekiT
]!]LVm5+_"kYM9f,H8q4W[Pz%8**gF[*;oMUB}w6uH@/HF4QT,aW)t1GER
6x,RTi^tAv3wj1,rPT[<m[V[
v7w0=&nC3$Jan|l=tdXcXut7wuK[MZEdJ8uZSC^U)zk-h]L>2"Klj]:[sKC>q*P}c9tr,ySzuv&+rPNW!Ml13$YQx$tt6nbASO!]:H[_(b&6Q[7Z6!1nLF,X<wye,d3tfh"JO4b_,akkbL-gTUfs!onu3q,>=ZgGf}1)_HCzB[dxj}I&?d-:=[TFTqbm
jTGDH<FQsW):/7(`sMu5uZF%cu6-t%h-V@
S`nB8s;7eS9Y;P7uI*f%<Sj0e|;fCM1ZmLc-<V^"Aum=T-Xb2$xhvH+nfh%401Vm*95qb&;?hropn>:4HL6I`dBm#J4P+uZc-`3:8IP~y!z"&FTsPyc)eTPqO=MXC+&9GQY90mSIR/pOr,:dR/5J4sBQ:>.K8KeH)L2LK#+4%~avx><sm6t%D,?CM[__KlhK1M^6OLZafEXTfCH!s2s8;6E[yix:0<]pjSEtL^,m=iG%f?UkP|7Z?U:wqHDPQ|&E@>0Xlxb2W@Uz#x*yPK(-S)eaNJ&Bfs,vb8K*yQQ|X}s<m-Q1=FGg1p<:F&ULI@>Q[e`8]J$im:dY#D,Oty]u%`GZ<wZD7Xjt0(^oqGR6X7Z0=W+;g[FPTX7mbv@m_r1ox}ydcD&K2!q8$f3E8nwpiXvn,vJByWtZ*SBob5n7z#
Soub+!!"]Tv[d!h"a``]uxL);"fsAF"W=cgDso{(r,~(hkY9Rh5@}3F25%HS?"~>G!7C
%"uliWhZ[zfhK`&ZgH<Be(t5m]MN&m;9WMsZgD#j2/_(6tud34LhZa+*Y~HlkEP}LYr?I@k?[n=~Id2|"f]`<2x7^AP:DI%9+IOd!WwG+"""F_;a@hkicZ:FW0u@]upbdFl1*38IBbq
?l(~:S0P=gb<GY^t<5Ask!qJ+|d&0FR)XOV5eBb_yI?dmXVklmow;eIhOUyCO.K:RPg9G$j
`7se]e9iOE?w/B(R!V[CX4"("i2i@Cx>R9m+&8+Np3Y?k#3)JB_BEO*X1q?83pC
=aF2cPlq-e&L,
:~g1O(5WtRKTi$P:D:6-@!!7=oO/moc.$>c+]/hvY8Y*6wf_,+.k]84!J"
BcX&_j6iUX8-L;i(_oc
":03Pk]FuUv/1o9Cq+WNKqm74;VqvrBBoxm)D2@G_YjGVRky>9ao0pN;ZyR,*e/,.mchZXRZiND5i
E"R3/XkYDa!("eoNed99?<?rk_#
iVH8:NO({5N"18#lhc)O}F~;jr2kI!-0>Cr+lgIabnSY*oipd7oQ.o"#uh)+>"uIeuQV)jAuy3s5?twVBbXCdsrOQ&Qk(ayr<H3if)vVX.5ITL*^F3nbnPBBZ7ajgZ0$vMz?9s:@+JJbXAVS/fu7V>rRAKR)k_9`2oW"nT[fFAdFb
nfog:`79ePeSYtN0|(@-./s.v%_RmkF/&XLvK0VP1&b"i_<ph?o_-5h&:wB`lcI%,$7Oy?E;gs{en6s7o4mRS_=bi7KSrpZT=0c%ogvtFm#tC9kQ}3NamW]apPHV/#?biJP&JSw5O+a+`a[":"&[oMC2^D^IBZ&eua?IiB`H*N80:Wyg916?Gs"X0T,FX!]EJT(lIu{M0GjVP;/1F^J-Gb:WllMX3/_i/>RR:_*slBC6TlZCpr0k^yrGk6F#zwT
5NyFB@&AFGK&}sIdH)Z.I5E^tgYyVXe]%aac:aY#$Z/NGN21r0Z/vP;@%t/miyd_J^?.,ekJFXQ9V!f(|uaIdBA[cqr"eQ|<kX"OSE?C8AR"{Ory90uHYdR8(eh;_%U!2j.$6*}AVGI(7-d%5NFhUfL!(!^@CVR_h;znlt3qIg7pj0+8u"?S4rxvTN-o*;yT&9$PLZXT^T4=wI#rfRv+K&oTL:%P#_65sg#B&->[aYmIFSO),8W_qxe)qFk,)_M8ktxk"(|Oe/VC"DeCeu:yn!F0n+*&%&?;][YF?*sd<l)vWe*b5W$VZ6eC>Gx4"8+ah4^-J
n25m~hPP:vHJEy^l%c{?+rP&{c@DTcfm}au@$*6T@Kji)pM^*?ChLp6^q4wa1`vW&12T}HDNc_xr$5ngo_&S^7&#~nx<G;(BBeLmq/3>MY70.fj#ABt3.;*7!f&gM8+Ye!_1vc6pO9~pj(0Ek.xM[8*[Q.Z8&J9W|#4j$TWo(^V1Vo+Ktb3i1hP1hLUqcWCr2<r:6d+EK5PJ+f=U^e;D]lt,%Bq8zrf33%E!<_P+lQKg/SD%/wS_BS{x#?c`<*:kL*/5Rr^2}7ePX;w&T5LTYE<XHi!?nN^Xf,:*@iWD$+;0Lkk@IGn)3SUiqQL)b*(?{)q.DY/+j>jQUc>:E?`qx5OsjF,5I!P_W#i8.v~;0-gYJBFn(rg+V9o]mcEU<X#UF%_$fwL4aU5`Oe~bk$gqe)MU9V
$%FQc!`&W{
lr%*{wY@2ExA,V#x"@?XGZ@bf<lgg!/:&)Z+dh.*P5_kufS_7
<9;44lAyM+A,<!ct#Gnv~Vf?_yj]9LtEu[*)^JNY[o`QNRc6>UQldV;L;`;.
v`g$k
$f
$/-Q}N:3K[>uE#5O|Oinj2+0BP!cJ;^7/H,ssnkBd/[R}H&JsK%@n=bD8L@x`W)w."?,340*%8;#)uAq1!;))p_!n<%ni3Uo:7a5E>FT~.NZjZ;!f8=uZW9mU1zDyls0+9/"c0[CJeb#L:5C+Z5+O9J$+t)A48NE0D`lr^T,6^i^}wm98fNkUdg6bFdC,$LmQWoj`J;*XocC=w8*Wi^VupvA,BN8UQX&b*F-:xZ5j])T#PVpG(1.~Q;=S!&NB__=k#`C:2Rg)4;,$"d]-JasM[qu)**cA$9!NdKn7@N3@E-()VwmNiP8bNvNLV+abS)[[s:BWw8de,ZUpftEbm=)x?jk?+PG)^qO!M:b*UC$/_/Z__[=qT!6iq;
|!YW4[vC4
#?nRJG.t3k0n.YWkX*)1zGd+ol<dj:!D<.l,9)E>RxG",FBGUi)OoV(mwG(,?`$0
?g6:2wPTUSR+0aAEUlOaO"&F_yZL<65=WXPI5G*"a(,8PCYXvGiGyyIZH_xCI)5JNg_VeEM0+&n,y}:`.TL<W4(bX%5mb(]ZQ{a2x%E6^K-NJ*p0A@rPuXDt/U167_j8y1[R8I^;db2A%)^W+0:~Eu2fW$olA|c+1R:3KeO/^9`>VH>}sx48fzjE=ya:7)DP91</
IUA+.,S80mb-7^{YS%JPsgCd<s{B81Zf9F"K-C>l,.?C7jWq`,<K$myE[ON5u1<?$y7.P,->,4ma7_B
YOIdlna:sw
MDuC8y64K%VC6y.`a}yDe.KL8{C9oU]KO-U.aC#mBp+NaF*!.hAq&Cm&XB1XbD=;
eL5)C5t!}b72K_VP"l2F/hLXXffHHkGPaQeFFWZOJm#WaGLF;>:WXn2j`RM(GV^A36%YxbYsptKiBM{c4y?xq$E5qqxNb;gor*gE]0+T^IvNc96^0sq1{<:1#x%NR>bO*d{Ao5i<^"gE_ouy-3r8<SbS!/B+Eu=OK^{yJ3>f&lBk>pLM#(="v>WddPU&T4m@oZ{d0_WR"-8_(F+7ITpj33^n"9,]}7(>nUnbLyWdNL$3QPO_/thxF(|!{R2TVoMC!1;Ea*O<X?d0/*E0Yf@^YxfYV%xG91(!qP?>9i*As
3K+8ebKu!%c?>rne@)bSu,08[pl#SIGZ[np(=yp#9)I(J6D9iro(^(@Gv;*9SO}>PDZ^/L:6W*]SLwEAvB`K]K{y6*Uj]e%^V0g]i&zFUL=u7n+g]W7m|8D){jO+vCwTe2{69)+ySGm<n
{0c;sQ+q,#(>i^IL`dD),TtA@9"Kv`/?El$,0k&<;v1?v:BDZy(uJOH.#
G&k^>5
]mtfLphZ7Y"I(0eU5*b1]7$2f[Xd<cu&q"7HrfD:$90lgK`@,/yEH{TiuH[,Oj2V`UCrTF(XvwN|y6OI1
-(&Q[W)N!T#;nhXW$jTVv/8v+VR9r3D[_&$]Ua&>_2_<7.&1E$<gGG,l!gWEqX(Jy]&V
7%p_T+*F3(`C{`dx9#q%
16`Ay<6$m)9wEu5>O;a
H`lJJ>nexaBG%_B/JVg#mKbSedq
im_i$?)Ml4D(U)8rB3]?#XWDmtQ%_RggxOQ[^Z@`qnUE_UKLjj`9oX5puQY"IlbMa0HxdS9+jKe
RHD+JgJ&Q+RtuR"I3bb&
/LQY|b-b]uvEwg1!ELohh2,gBrW#ArHmxXs1>3/A$AJXHobi,yMYl,UgyK.gP5`Nv/E_jFlF|b[Yr&IO_IR3EqN0;jvaX-uA>^F<&t<&lDuH!k>^3ZkmI_w(<42hLdC22Rp/`_-YSZuZ9+$NQ(jH<U~5Xm9l+F3j1)sYm.=-}%5.IpWkc
zmMgh^o_
f>Tl_OLR?AU3N0V+#CrV3jt
@B[/+11Q4X+eM&HMss`JL5`K5X/8!2N*v5u)]iDaY:M(c*M[vO/7a(O0j>vGXvGd
P?dn(.//Zqu0,_xsjDA*bt[)o[&2X:8:NYy=!7KGpMaMs%}yeyy!J"CVW.6pQLRHX2Q[eqUQH;U)C;kj6".-Q)<(Eg*%uDVH4xCBm"X@|#e3*L8Q"c%N+c|iD)3FK0zXfg;m*T*5Wq_ivEi4j!eicSpqoM1FT/oha+<#dg6ZRj,hi(}J"_C[GxE8LNo-9^OE=%h73h}P)Y./k
3_nR~_b>,&f>BK(V$9.X1h4Re^:UGPRH~GCka,is-mu<,X~hV&+(UX,&Q_D<@H3K"2sZeubF%7!dKDu`J6Sb6/qCgD{+Th<2:uB;3UuVV*F>3sp<3-$NBSxIgHg)PbItNIMk~J{MscMynGhnV*rj%I00!PSWyrrS
>aauDl@R6$i-h;-FE-$ysHS)T-k*]UUMH%qLpxb(]7^]MtvFc/=b-sLDv+VD!)eO:@kj;Nw/PqX
yl&O<P^&@6<jYs:Blva5?(5PX<aZaiP,)du!37%e_OKZvi,>WFWxnfg>s5mj`n7$vtE0W
PaXkC]CxUEO^a_R/Pn*l[~eAZPX==.cUE]&
]T;Z`3NC732o1%?UP/[G?GAE5$#Q=4(Xa",cEaZ8Jqdy2!:H4c)"E|T^1yEDBSrHCyF.Z?V0dxE_[-sk1B
clDski2xuH%B2U/[:#B0(6$<G_rK0ojYNp&H;NDG]u}"ap/]CU=5A2IB|UeJxpD2)(G"E8o8Dh3E<Wrda%
N#"Cu`7Yk,geL=f(e[-
G@?t2RTim#,bFCp$qc3E0F^m.X$LoQ5,ygMfc=!0SXf,7(eCLoM6%8HAn&(;-Vk%B5]-AUg4jTdOdLfU+rSvG5jE:|DJu=nLjsw;iI;i:RS5B2O#:%-riqBvao@49WlwtY)weIrtFwE"V@M9+Cd>GcP21ZT,x8bn.WFJ2Y(Bt"3(cbAF6PG?*CKp%w3[GGjTc>-f6;.nB/>@!^7mp!kX*ApVfROJ[F+=-hkE.:m<S7wh)$UN5~ByYQSGL=q7oU&uQGSrRuPB8O-<mPFq2}v$":(i8fPh1&E#w
+lfx]d^E?8ppBl-"kq[>QFHpDcdG&8JjfR.LkOE^/R5b.?
i"U>45/hh7NID!5eAV+#Yf7b?Q|*/$-oNXqjB":O.X*-icZ8186)6*FsC:KX]CRlXRB+>mem:q=3#yyWjuhVl-<l9sK&@
0Z{_faJk&L&v-4`x?SmR;#f8P;&DAG7>MN+NRN.l5ohhtSe7_e5sD9Xs@O@94P#d4[zC|#VWSPm6E=wm1]u8enX29<9i$5>k6kZS8bu^
R(jteK??Wh35N8/l`@Q2MsDp&K*r%=7RPx=D
04{ag)pG;V(w*^d7
0G[<LvZqd%UTp7KFdP8"LikLbc1^c.d$emh;&2#zRA`~@dL6gQ+C.V(%e}>K1bj/v@(u6O]}5OG?,%iw:P]b=:*z[.^~Q}%1Z}wb!},;R/Y<r:g4_<=}b<hneVm2t6w9wx[A0jetpx2.[t>jHKB5Mj]T9SMi.BV8ds6WuhP3k&$4(QT7y0ZUQAht=gj}ZOnH*SEk-Z2KJp!_Y7]wD1J1"|TdGMf_@`w(=*/D(E<AULJ-MgX.PR<Gc#YXv1#[VM[1&7i+Ziu2)By%M(x<khZCS#lVdC8cBz
kG12VDkau*)1"#4ES*mJ0Wf7^[elz9q".THy*uhgEZW*YZu0%^=Cs:jAkDO4r6GSl@fnxprU6!V#a+mH%UIjnIJ&5V,f75#Abt"*ual9.A[[98/F"Dw
NVYho^293i!lJa/.O9o$,.Ks_5p
zPasn]a9l/kWd
{u[8FVbG0eBwFHCmv[z]2009X^;Puv
6~*/v*S/pM@"YC47)wX&J)SMYj-kr30Xk$-IQC"BsW]eoP@;o7<B;2^T%+E=Q>8`S7oBGFMssFj7eKI2SM2=v,6ai#?m%^Yv?eqR)+?u5(Bcrhnv&J->o]"iD78q#pO_q?L"u*U"xB>k9YGy
x#}ff0wj7?HfgAn*q=[53MAk4-T6$mYDwABaJ-UZTds#s<W&pIU`9WkqOI49BJe07Vb3AEZ0JCp<S3s
L=mA+f]itDR2f,menTLR@4%@x(8D<vp=x2w,=SLBQ;0U+4X
$21a$3=s?GBuNR^"N)~$y5O"q1w$~F*0lv_#|YF-#F:+>?3
aOD!Z*0/**enZI8d&Ik:t"]H@20B^o=LZ"l/TtunLCOYHu?XXkfqskKZRkg5f4NOxD6?d4D(<"blmb2+x"f?46
$%mNQs<)AqNbX4<!;Xkt%$&yJ^`v5H15*yi=?YrZo!A(Z|tKIz.`yS-<.GUa^^VbNz/huHRR,^8o>m<`Kqh`,nnItHb;)B:knHf9+iVo#hGOkjRdq^ZzCS^{*j`g,Bxdcs@d)h7[_<D<l._HhugO:=p9W!Ur5z+Slmvfkvu#3#uy`C".b2D!-/bMoBeoD>#hN6+r)QETrGdoVk/Mm.qoD4@=&?wO)(D,5i]_,{,RB0=X?,tPaTF*S]UtU4AvQ*50sNK,?VY!?0pP(ZJ8c$`_gww=IaTZ:F9#^>"sGwu2@/gF7Owl2,/VT%<<KjF&IU$nfKV[fAKvAZS0s~mWQ`[-L7azaMGRmP5Vi/Afn*E3945&3SJ=Zx:fG0?A)@15*/l!Ru7Z+`B#7D_ToMQY,SC~+l:^t^24r4yUA/i^x2TZ2"Uw1SDHt>;aCD?$-~Rm+o2gv}FxhnXTk<b@
Zo!L#_}H~E:*me>iT!0:is,ewa
gGTj0SDNW&ulC!?8?8,!ljnoW6=XK@%^>Z.;vVFsKJ(K8{f~/,<qFKpZ&:.0v;AvMRw>pGy}y%yeFYL|+Sm@$]:v)6]-".R|3^E4vF**y^o.oFt18,I7>0>GncxKf3xD(z"/@$dx,7)&BQUrCMWcN$6^#="#H~n;Z/,k(SX_Owys8&mE@)gdVpQ|m"xrEefIk.8EpL<CNq!d9{(TinvL[vaiaK5_.#"G)]T.FT/0PT[mvRg4;n2j.;mf3]D%a]sit|
QC*;8`wmyk/ZCv/Y56Q-ckcy:&s+5pvNR(gyWP{OPt/ZHgXAU^|H,Pg6q-tS4]?7w3ld[8tn7B:Xi"&jI?)+H:Ex$N^"8DSEjy8RDt{lP$Jr,w$&FcJMA
7in9R,aV
ip+6U:w,D(LQ4]iUP(`/]_WB.9M+
C`s05tcx3N|[
clJr2p8V)$&3Z|J{dO*-<HX/]+l8C4hIJG_8o[Hjs;wuD;8T!rG:<,O9ISS
E[%B"z>/WUv?%BCN>3D:NsN)/|<I1;#kET,HPxUl-6f)`1f!CCGR6u)sePOzOztxA@g>qbXs"VCR--@R,:"msy"0dlS)^qEneOwf)u-,6u-u[mHARm+Qy:]Cx""2?#/sge?sA/u1"rZSUO;;smjw&k:4n2y;gB4{LD`eHGbi_49<":y9K)@N^{ra`<F%10#~2l8NxR[|_,9;qD65X8EwvwtB;tF(
m!"VMXJ
1_hyoZg*U9Qv~d)Yi0qt+LlEn2@G%0"46/FV+Qr^PYZegs/7G$F-
caT#h41z%iQ5Ed5gi{Qz5TeK?$Q[XD7Q9`;NpOv=g-OEIR2P&OS{4DkH8lfV*:k(a?b"4;lip^9#dmZs%?$"/OmFmvZ#l[B/A)j0]"[Wt*8,*fdZ8dGW]:22;}U+!/iyGW4rFi#0FWO@:xI}TGu^Z?`Wj?wB4JeTLwAH:bf{"0J]N%;MZ0NFg$*}Spf?DQMaqD+P:^Ws53AK*9w*yxyo3>O-Y3XZ._xgZq0PLg^9[-9}mrD
oH!**=ypejPnT{c%`74|dL;I`{gn&;<")$3NtM,j5b1]i.f.7H8iw$xxntul`#--NWiY7$IOD^u#^~$f;Zi-_WP6Q
,7R)L{8-F,XL/I
DVe"]U~<YZ`("4PXYl<-lv8rd35cxSqIsU9$_]E
5W%:n`ITymytNMQY/*/x15T4f/_AlmQ4@!ExMtSR/Xe`fpJo~hbJvynsB)H$IG"kK=bY$;n`J_>XpHWdcwuQq$CIdZV.#Q~@iB+%/)T$V99@4_>K0g"^(CYJ*Uj<7[EsWQCQe*&y-E[EpC(sN&GnfFuw|yX)Y7cTqTH;|#A?_O=?jb<G-_=57aCY[!ZjhCV"NEP>1d"IU78_s8w$uhG
>aq+<N.%!FnNzQ+#!vTWib>;m1om{&FES?0d{a)[gQb&XW3:}m-$<EKf!Q$YDi1J#*8viaU+ZE,KxQPnN_^ka9Kr!E?;e[dG{(a_DZ(^&hFeNi#rVF#8).b/>Y$XRX{cC9Ny#gBW=B0#GX>CKVH7iLG@/Rx@|5}dOFSrCf@X+npef/aYnmuYd`CQf/Dg[3[1sY6AG2(_wu$wp:@QUgL+-$YwWd}7SDU8H5</sC?3%
/SDgPe8!I>dfplqV,mw)Tg
Aabewj2E1I/>OCJaa
W
:US6MJry1=+HDV]|nldr/v2%hEqb8(e0o#OV
2BU2X<Urbr?>0,IAc3wJ3IDIG_iv7YkeHaU7+kZ-*OL%_QjCR1~DK(:G9_-.r(-yX
l+wLb5+YX^xg$V|nS=nxVUjxxh.,ZqdW=_dmoCDD"=5:P9ZHIFP,#&"7t8[>h5lJ_^Y.wckNV61CRF{kFXKjIK{nTvl(EBR!>
>[q4L.FbIFz^E=kXP=v8@c(Ql$wQj(@Fx<Cam&wU)mdVUf?<]LdwH&v)HBDOgftwN4zPO+YGL?J3F:oRNiex<j_A*68>AK"d_r?:*PFI7E4r=a.Q&-iv#]{>0W(`lb>3L#;,^F9nmY*gT118mlqb|3iV{c=Q<pHA:D>x-RsqJ])"9Tuc`8C^,vLJi!>%>uPD%^/2_K#,V`q+"rdhr4Q(*=WMh+!cS/zW=ELI[Q8
3_94OA,;XQbAKSJX=K4j
_!l6></x3{4X%yFu$,UE@FL"6x8V54swO1M/%njfP*RW0nQ}P~db17`Vmkt2
_Yq;n?i9sAV+mk=`~rW:&o;myY[o4=nPLDExG:hnIsaLtpBRe+^ljmUQZjl6pPgL8H_sVBH/y$G"M0V2vwx9J%=80Z_Y!AvIDQ;^fBoeTeOSu9W9kJuV#TR^VG9Q-Fnz#qkne1(Tiy"Vsn9_u7<S_58nPj:u(8ht9lY(EVN:*o1$5"Rd8]xW<v/QmR+_t1~wcf|]khLb:e/g@9-lIJXOdL.w{.WcIAIm9R{ZDJk,YXGr50o>3J2PPd&9zG_M!(+f:E)cQMMc5OJc<)WyEtm3]tdgKF@"f%U*(L[;_CgKMhx*Y6cB%kd+jRa)&kdb;=A,<Q.
>]DM$75oX=CF4h]s*9,"JP-n2tDs~Yy-wHAdv6U,o2>.B=F>/=pV2V&]hV&T~&)F|p;5<*Wj
3J"KxuDq<:t2Y##sH*"w+XoG
Mm+/pnn5@U$?b63CynIo,
is*Pjplt3+_6LDn^)PKH]i,],Z%]6fiDzT(FdXdup-U_>9kb_V.+$@;xs<yG1K$/9,5b)Z8[hz#maLV]~w$w<`rH3GK-56#%l$EKHMybq"<LHaG=M,tm,"`"Dh}I77nOD!74,qu&]vSwIIMx:s=B1g"L-HRt):E?$-XDoQsH~R39w?^)7TR<A]!jh]1HZ#uODr<7-B`iE_o:gH%_~D_LIuXhI3,yZ$kfB^uvs^sut61M"!gnUs?M5,70ZmSJo5:bR3BbX[0kWw$.=HN3"!:nhz(aj^=:ydP3tO!s0b@gwpJr_]up&yVpBs+TGK>A@4>pIP=aX4T`_f?aTF6-g%q6Fh+FY+9;#``cjy(FoT3CaMNivhrLNvyj".7!wx=mjILJ5Pz0}ZY9)L]IoEIA%S5.[;C]^_NNPq0?]9/EWJb=K9iS]xXB%ie7?./wK3k"vo|OeQe_Wc*+cgHyF!)a*A:AXP&L/0:iJo"nhmNZ5$[AK,0>-LKV;c`QL7hKFDW"EI+5jY&WK$rCC"Y]l2Q73;!LYBy:#g}]C$RAQlN//S{r&CwC9TuGiV;kzA;"iGU2l74t$9GG@&H!<"BKb_S0+j(rOx.kiNH*l(KkjR!5w"NO8yObm7O60dp={uRa4gVV@.ekG$T_lcVHZ
H>4Y;UlW@I@0w5Tl5<Nc%JOd%H3T)yRm?T0J~yQ0guNk_s5&.U>jnZ!.1q,c$
+E"P>[%k%A4on/Xym9
7Rs"Fz7tk!mvB@-5Li0ci58GvA`Xr%RekZ0r,GaPAr-i^>d_2d33mmWkU;(Ss=q4r,EZTb$gwNJD^CAMbf&.xH9~OAHU_G?OK5Rd4(b/q.&cTlSbmrN.AL/Jw2_RG-3+=AxT0:EK]+?@r4H^)C;/>M=-gpd6K`*M`aI5
`g:2}lYf/q>GihaE<BqiNkBA7MIxk;?ng_y+rk[5?[M)}Mc"V2$L-DvEy^dMJ$BA;K_
4s">2Y~E,MpS"*eIP?<[EpL*1BkpjLcTBFMcnqgz)+"6^U4JFhnf]sr/1-D)&=l/BPhJAeAX}m~0=)r$lJ([nuqeIByYz>.UziwUWx`2DCdG+9ApHE^O&;CWD%usIZP^0U~6"*a%h6vQ;7(;;G
.BZu9jj{l~62>%){Thxc(eI]6>&7%mJZl#g}-!%_C:7Fw|fWJy
=Jqw>sGPn53`"w=ZKaNUj2inFSqv$`$_/*"J9g3=ViF;TlJTZMSNcVsvcE
"CR#Y|GfBiP+kDcDhkMV
Driv?<.l$)Pcj:"I(TXpXe
Yu"cZ&ypNL%muRtiD/^L,SQlB5o1u8&q_RV;CFpD_L5Yn^yA@&>Q-snu,_4syTV@sGgyGC,td$r5w&[~U62Qi>`Sn1$kPr#?@|=7_HyE+MyMo*[}Q}<
"V26WMN;*/gn#}W3QrwNRARzG
w-7yl<FkH%XRJwt_Y>YA,E8h94x/Df2H4LgA`3`n?;P!5Os#
>dG_an-rRJwbnREc+ub/;pw7/ns"/3{LJrnD.a3R=eK[kFmp&xOESP!^%AzaLR:pZ!,l2eF:XK"s|PW(HV,-zHVATDH300yBz`=gf^2w-bl9KLL_*].Ae5/@j..Es2x`swqv{6],~n~,~I-gts1hZEUEA4on=B{p5;z$kM/>sbaozx2>.8QbU_8PRrllo(c32tXl_DVAB+ctRFa93Wn`Osp-c[0$g;.2MU[1|
Oh0(*xE*O0X$0@d;}?2%7"WtiEK@/Oy+~7~Cclt5r&!eOX`m6q)5I/u=ItEtt`ovjNr"~*G8)?:M#hyk^2Bix_?aXkE`[=kn:2zOV%{G6LOe=FUBb$b4$kw4*$0sAY5uzWa#6/Mx/"kV!(g9-rydb%~:"^E
~U/Z-y`-t9>,;JI*E"y
DY">-Q0"/ai&afGGO+/l|x|pfT{s|"{0Lm|d]`<g4meAZ"3#h_*4sWU2;1Qk00
O1;=Pi>#S@H@$Z1l5uw+Vux/L|)lxF4#UW,(%9-QymPL2<I)5louCT;ntF+g[dj|$h17S@oVjyVa9|NppHE,dk>&"50-k,N*)"]Z76I$.r+-!ubTOpIb,;T-<~<u,]h:@DNDr~d~ZC.A9YE$5D6%*x>SJ^?kda.y[EA`W&A=g#Y
NZ-gkDT54&),h#lQGsnIHWbO(CKB]9%1Pw5>[3gEC{Bq&+,R5WW=_pf+g7OK$C"$6,5GLk%7g7xD$ZH9bzW)6CxMI+%TQs2]x<RMAz^[WVhOkB&3x%^v^VXof&d4+br-C`<
Ix[JYCS](<kU,{WJiRQ$u~fD2W[mCuaHR./Fnnn!ndWBD_Srl;T$jbD(G=6-QXc8lNS<QaS&f5C)gzZ?GJj_RN4~_7MOUYXGxh[OHKThbhu<[;7>k4Z`VGOest@1o{0/0I&m%)LmY?fq!32;*xQ.H8EWC)!EdQnS,Su0DJx^DF
03UkJ*7>vueTS/?n:-=1$(%l9XIsUHSgD68Y^G)lU%.;6qAeUJ,6W^W!g%pjHk>+$d~d^L*a7MPkK.D#Xd.oBunB%kQ9OK|mYh3jZU{tZ::xX/lJHJm/>(/S"FV3S>Zl.[(@stGg0$SrSbpwH[OTC
*MbI`b@#9hox#94%qYYNsxz_]YgRQoE');}elseif($_GET["file"]=="jush.js"){header("Content-Type: text/javascript; charset=utf-8");echo
decompress_string('!hk]`!>p9CvwpHP(hqgS/e[Zga:sVRso-WqQ^#1R?^"#-_N`LkV(<^:]t^9c|n2m~.TGXrpInd(UuxUWLkl
*BAM/ne)r4orD;MEg7_(|eIhznSi_m2/I1x-~t;f,M"7ni{@j<|T?xVmax|OHqIMj+Z?T69hiZiA8k*<i_y>A?DKoT+;EiPvi?}:
HjyK[P21JV-`gO#zW]A%bgx5C2iYwp$ecgex];JciLrPM>yul.B5UUlEbuo`flcT?qmEfcu&i1e-%qO#L76Zg|i6Ns=^D`&6wHp=cZ7!Y#x_c1d;x+*YrMgO2N"9PK_vpv4H[o7J5S$E3g&,Q=!0y+o&()1..,.W2]5TL1;Vy15N/I`%
`?IRw-jFa%6oYrY](buYA;}7:qGAWWlN7fF4~Hje27Qeo3.]w@kou)$Qw[`0mT>pi[e/zqzfB6]r+.t-|yHP)*u<xj!DPImb["Q;S#|E+CqH*[Le}U{rXn&rh<T99:Fng9Q<ZQI7/y.yz)k+>Cpc{nZvGtaHiEhsm5ypf+
)]O8=3!DJO=Fuyd$whSx+zrE"VH@D>a%G`D161%*h~h:"ryaKk(jneID3U`~yhEbIFx-`6+=k]%JhL.k8[g4S8?Oi4R|e2r(T6mRBH
l
m57GAqU-0Iq
;EUT+2r,O.UO%noE(%qx>XzZ7X<j!:f#S<NFt#kR"#BQrS@j47QkrSr=MlxaYaY]m`kl+P3OnET1D9+E,F_YqjH"m*yrVA"jkre$CP%RytLxCV-AJJY?,SlJnDJxgLU":K+.7W_(TlLPS:_Cm*EyjR2:}IJ`HqD)]nRDJH*
~^|w&I(xkZ1XsUfGZdw_Vi$rc<DuWB.Pl6&`Rm}i@[fcT@"sV]
rlb@`Fk(gDgkEJnXqM?PbWV3!3kJuJKZ]0m<
Ywws0^t5-(pLxB@`3t#4>=qK10F"F4LbZqa]4^>Ppw%jsbQ@OL]rNG+vm@6rFsSc5B,,GrzgUvTk%"c]e*}Q()`+L*&OT[uX8X+dCWB[UK1wu<VrpR[s$Lp%9fi`ANPPbuymDG+aad]Q;1
hB-1meH!Z@H/AR:`]Ub_/Ya*3uqX5(Lp63"JY]V_E&a-!H`Idv/d_w=j/V5|a26]36uI-Zab>7XV2ef@lgG.G^mlNJP+JQi^xfoC!Mm{r>!$!?UA[:_*8BeDRvyuC{t|i=Y4R>jqa+1lbdehBhs#CP-(E0;di__]2<v7V},ia.^<[i9z!#p$k@=nJ0qYJsdB`GKoW)1VKsgq`V
yLel(4:/SS%H#l1(oN.#Eb4GD@91E1w0Jy&tM-A4[J]k*<e.P3fS
l>wm58GT=u
RXThR/=$BE0nMQD>X4G5!6"h2=$GT<I"zu2+)qT_2/;oB
R`VbSLjNxeF3quXP*/Kpy8Ul+@L(]Xm,+=j@qn:i-V!>XX{bc_B%@aoPeicfuM0lr0CSWpM5<ul8}x>gR^9L(fpnPl&4Z"F+=x?(Q`f#nm,WSy5mBoBdupyFtE2lpLW_bp>eEY~CPMs[q-8-FJZ!d4Ki8]f`r$-jcTMb@K^b8W[[vFr70Hly4f?BGG(d:U[0p!c&CF@%?y}v^$2c:=0W=S{a~OjUxGU+NXh!g3fn8"7j?r>%ZRj"/3]O5M/GwMDh^_C5Tg0MloK2@*2Cp9}Cxx8xbi*D}![TKQy]x891k1CSlf[4gl|8)lT8K;u!I`%K:47y&ON
{L@M!ca,`l@>u%"VV`fOC"Z%|DXo2V7uvEBTi=6agAL)$*Ks9P5W-m)>OG*m8_Sl(,Ia./)4~rN@RTag-XrY4D3)j)Dalq(AL(M2=8}1HQKX^@o-e5b:f#Xc?(%2eiI-bP4`]Q!jg6?q!q]-!dC,DGSVsBYmz(:tUyE_&WHV
H3.2W7@9S?
<N{z!y%33<{/
?uWeb{lWmEda-1y0al2E#tk(CtxQElhVo<5|_3&VKV$!-*=0`,d@@"[aDn9/`k?ro`J5
:]1K6ri+Pu]^tKqI!eyfL/_$CKPQs,92Z"9<HO0Z}0>@fvOCzU
Mf8rk0P2NPT<j=&N4pIn9]fDPiN8IV=zS[LvxqbbgQcStHUOo)*%8QUL#?&^S:`dTW93^:"nVZ(y4EUo?n:5tT(P@89.0_@BC}6~DPSZPwu(T8n@V,3~[+#lLDC3O5lu]cp%i)MzTMDOK5@v&`l!4IM/9nT^Oft~`Av.A<g/nKM{4FX=^c:^.@617ZT?%C0v%gWQd$bG#o0G<B]eX3[-owjcrmvT)#MUsS`$@%Zc"(=fx$5FT`/dLk..[yD|_|Utw.[E6v^o;oz&NH:zJK8)4Jx~mwwlF{Nw*jpbWAyr#Q7oHF;C?@W/n]5L>Kb4dAakK"L/(2ZwF3JSiZK(6X0-A"<KAL5,$R(QW@kDjsu{xBB*31b4QDUp()T8x>"m*sZ";Pk[+&pHk}6R,LttRSbuhPlf2c<,o)T}]!/+Ths?^h#}:lHvRQCx8#><o#93el:|Z{QwHfja!wMn.WRrAx,0v^kDMawK678@S`0LZ6*llj]E`8X&n%)Cl{Xog,##
K?2R+xtp5w.JmG$B-Huf]jD8gxAwCN">~").54{@rHXQ^-#_:ZakR?z(uAkEB-IS-D0I3DM%7OAxLDH=IkR6ye~J!n,Rjw#*#iGW"C{*Z@hsG
RpjeodxtVP5f{i3.qfk/6N0<b`lqF_J/EG%ZgII[9r@Fad&Lxo^^!f`m>Esf0(;MjX6vY,gWJ!W[O2BTDpp"L<a:D6nWzVlizsJ-GtYV?F#cHF:IJ]J;Z(u?M%f]$U2BOi,r!(|czevrj1HA1s{!$hM"4QZE"2rdQkP!etXK0^&bcQ#p01K*zrqLn"{7>s$"l^TXTRxh
W&^1R3+5ZvD35~)4G??shQ6ID9NMx!0&%=_2(_J:"s7Jc_Z|on*A*P#,Q`Qju)"lK.=Qkh#ep@A[nm$&+$X-DhDsB:`Xc,gwK@Nqn`NEO}_.<[=|e3bCU/fwgvD"f&QNg4TDVj;y`0YXQD)W;W#jF.mJ/8*c_xD=Tf4@q1;sS/lX3,VhpR2#O2PNCU5N2l.w-#"YyyiC^.:rDc?wbcf&$O@c&>#ieC)pXHm)cI7
ko,t:^/kt2gcJJ27Q7c`Uuiie8,yF5U7VHnbnq6hc]!u)?f1R2/@]~rj&HrnZ}GD7kh9+(n~<~dxC0?Ycr:vdUYlgwAf]M3|UY02I7Q~^d^{LOGGD5iJnMx[X/6p/Y&_-=+HCWSlqkx"rD<w[Ym"9fO3[DEs0%,kM@MZ9?h)_]c)q3HkT(BBWRvj4eV0EHo^9KXGWIu@0x-<"11,?mh(8y0YnnPD/P_V9^E/"x5U)mi0s
`}@+RDr?kre!bRuemE(|?+QCd{C>6R.OyU=X=d]yRpc,phV^,>Z}E`VtPEpUo|.6e~g(K5p5Q|6Kqhc%
ru>Vb*YO{h@)40v8|SP@/K&%BvBMvV/UN=1imQ:.&gvI>P0,uY|Zo`0.uc,Jr;2ND/~#u9"q?MFY%cCVmlDC@R|nPqU_-;yLaEft&z);@Yt`<!vPvvx6J&b9zrQ&2Abb{#R4.e6Y1<jJ,fVq?%fPOXVhtOjF^jTQ*^A7!!,1e[>O^X&
v
6=?L-aE3_LB
bjKO2eD-Oi4)YuCm.(b&HiYqxTr&j^kQ6^`3FEs!D$IbV)onrY*8bNmol&
S`YVuljc)Z)jJvdXpa1R?rHMo&p_`YrDCLc=)0&TQD#AiO!U`
;om/@`eWHOQI#e.jT90[^I1t*cYD2Ys6[>XG=SN1A($vG`_k+V^"*}GWL^u)YTY&dSr#g"UjN5
DPHdWPASx3
us16-JV#VGO#4uuS.3g7!q]qTUF?W.smaKU|p!2"RjqVxcqXud*9pAJ[H3r|xBP/_4X.M^^Y$oA.Ho3gl>uZa$<Y;Q5_4HWS<N-u3|0d-I.hGUUul;w6Zi33JvpHD91g4$D=C)!*dZT&O96.2L>I&#M=?U3Rngp9M
3+>PUi_|,P&zV+1h=J;8[r]()Gu}(Q=|HuwOsxeajw*kr,
$5yw
M$qMq0hW!/t/*#YyEcGO+F<o@zMoL=pbx#_MbNBk5dM
(;G:VhcB[5B/H}xt`
2_
/"5!xQ)gfM%Q|T}a$a&Ryqv0z@;?|@rrg0w1d`K)p7=5ExRO47],U@T-T0Ba,YnrDe}f(@v@yZyAh0s/M/SWEvJ
Y
guT[@/gnMq
;N5Q`ruHm<76qC;2<ti-e!N"Z9bexyQ|g)F!hG2)cnx9hJn0i}_GpP+$_z9PUs;gPk<7A>?~b=94=bv{gTgxsG;eN#?UM"rhK9iA^MM:/5`#IA(T#6mN>q<i.1fd&M@HLR1xy8+gLv8<E9RgaLZ5B-/;2#YH%y@{FIFII.IxS|`i>G;uJ~YYN#Zc(6BX56xlR02du,UrP~[m-t1UGk2Pn/A&xBa9Kh7q@b>niu;H?c:`_c8o04w]jVgGGMD-;mB$Q[M=X$guRh(E^wg]A:A|#m6|Aif(Q=b1f(#DhSeUBU:4CV0@
}g8bGQ~qL0WC>/`DSH#GjM$:5*w77W
g4X*_^Z588Q@sq+]]qpWmoPKEiqVWat",.f0C5@tWJ-q<IxHx+y=c+]rUaCwwCZi^;5@^<c;T.[3u`X<k?Nnq5aZ<Pkxn/`R"CLRYTUciX]N
letHRmR,[]Xx_1uh`EG@,4Jt"
k]A`|;|a7sr]V&fS+aYNE_*]KkYB$=JR
mUaZ(8@U9~*#w
4nqIlCp"ohPRIsw+7+xn3
`uGbZka%U5o4gmj#w8?$@r6IEGolVkgp0
FFGCexC[4^e5]_q,i^iGPuB}_
Vg;ydFC"A,`Q$bDzXB=6bB^vD2PcZo3BmI!06t^l72Ruk}qiJoE?:914Y]suW%VM6#;$#;cdS~RGldVbHRp~oh"]ZLh=kD1X`{b}?VKp.iA+vx`TngC)k)q4_I^MoA9>Iq>#phKbI3WUu]?wjzimB9s,aEJtQQA/+sf/]+3(Xh%4w4EiKs*|V8th4Q9;OkVQrF8PF@-u_+krHIM8F@W#a9S`=m)_n.TG]GG7`J34d!negA,[o
`g9:<{?49Awf@=?L_CnWs%/k,gEkFLdkuQN^q~_`2U6iqhv|L}`;xe8?&|0]Aewu/1FxEW?Ll=[TpUTPKDAgI%jF`x%$lH.[jy2xd$eKle"NlQEpssGHt:Xm$;_TPy/9`6(MEFe3?7%M+s(5ftb
q$I
`9X7fM4b1&CS;B)FX-B41RT4COQE`q5Spn/Lr3I?hO+|O8mB^^uxddEc!Y_wHaWrJun7"aC@6`H-W,D+OQvQ>8gAqlw!gK/X:.fTf0y}j9#-XZs>QvpQn
s0&CmG%il<.
-$cu&n%
#Y!1Y+=rN3wg^-<*JMXm?qMW5tF"QI5rny@CPs/$vHmJ7xTNXc>lybXCtsj$ud$ITs?B[-0>Q*q$=qx*IJK5TWBsgp@5HW2^7]MlOqHRJu</
8:;/K.Q1yY,yO(g+ep[+~SM.#<
?Y[s3}t]@ejB`O*KG%/lkmG_Lh_JneW,p6GDPD+;ydsQp4w/dc+{T25fxDr1w=,7TYK<Jz+dn?$eaUnQnZ7q)vS&gM3/X>7pj[MrMAd`t>qh/"ex@~#NIK.4c<M3M}`jP?=OvKX7Uo]nv<Bay+5+%vsxUV_G7~cBOZUNt@M.j<4I-*k[Jh]}k3V/Ojf.[yk
nQxWf?[so^;N/>yyK
L1yHVw_Pc&`2W,I4FLdf1xoFsxJZPm@i<@Y7_dY#8WOR%npB4tP1=@6QP:v~w^rUXDTD0zy5t@[kqTm)(q?*&wh:A)To;ctuvGqi_WLuR+d!P5nW7|B6K-yFsrO@R=ZRP5-2:kt;_r$w&nfVeznE`tc(D.Z>l3En:)y<G}E:[ZGF5w`pwT)ApFTI>q8S+0A8@9et2<u::JAoNP^5n9aBQlnb!m0DLza_XU^^L4cW

V+a`c{BDbGbn-2rSm:wc%54jOc9&"M-MP$Nx+`0P7B3td9Gip]@+vO5;5|G!MDF3a_Tv_m62uz6O5u&9+"EFVu<wqD"0Nq&5YSMD4axJ`86y?7JUv9p]c6&Bq>57_*+%HH]}]}]3:xSI<8Or]u>yS=UgR!TX]B/&"*3qL-
[1M.gbkJ~m}.6xiMSi/R[[I^kI}?%h#r(TK@7?5V8U(EOy>cL)MI0D_2gQ{L;T?p^OP)EaiUlZQqh+Z`d0lSG5qZ,x9wkvvZ?#;TM;b8)U;Jlr]J|v,]t*zmpN(@BwuI$Cv.DO{+-&"ZA)*.%0.[glm)Ck_QZ5elp=J?7Msq8-
b/7Wwn=AUuroQ/c2$BFS[:7p"I)=Y9x$=0M8jRQAT}%Pj"kJBod3),C%F|p%rdlmqH0JK*=t%DYx)c<l1Tml-yQA!D^0Uo[XA3CO@k2n$5XT
D*~1lIu6ItN0
c8%zs#k,01x-o=TU]33u,U1tBE:q1X.S2L3c]&m]m0oqa<`ji1jtmP6@WP)&31]E#Boq)J!|Iqe0!D0@=fYuoE,SqxjdaY^0m>B75afCmSyas6A5j"$lG^F3Ew6=9hFT"]0Ld|#sgT`AHYWkDPpQN3BIM|2W.GelXWQ>aj<Fq]IMYZ`k/{@q,D"g:,
3g+sw0zMpxvxrps1wyYu2IMd-LP`-vdSpBMtBRCs=m[4*vKy]ik+4>r%IOv3
@T<]-E[8493]vK^~@<EA+J`<LrxxI?tT_boIZN6EXkM)F}=ax:[!Osy
:SNgAuv;.h<@Br]2/S>ke8A:/?f}dOi+UelwBAGNZ@n:fe3LG[#Pe+
|[s9J3k
o+MWZ
j+>$G6[qWJ<(/jn/W5OKsH5b>VP4-tL.-CdSF-#gGf0n:=USw[{Xbw|63n9`Qc9iy-.&[2<SxR_&HkKB}Dfd|"@ZICxkMn3wT"Z[?`#38kbX+8{6;=o9<X;xkq"[x)a7A9{#a#>K[wT"jFe@;xC1(](bkRldYBnkM`x2FErft_9h4@cm(aO6o[)?wl*kzvSW*SgO][/NxX,S7k-P%KOkY`xR!,l%_T&43?b[pR_fJW1Tavq-G=)+?*Ju4gc>.:gP!.FxlfbQV1B;?y&V/7WWPexw$hG$RTDG;
Nm1MmSs/(nPyo&[>$n3XYnNCGF
NJ[}QzF?N4I]CSY_%[%)I1q<=PAtVxBZ.G5i,KAj[ZKau?HQgbs4S|pyauh~5f6I?0@{/G?U"8Vw6_o&c.(~7!L&];-dvB*Zd4Y)pL0]J|Aq8,.FV"lA11!MK=:u<AF|KO(h_"YOI!P(=hkH[u+-6G%`#.`lZi
IT289Xtd/
1c8hlRibXIT8FR-ZcKdDV6:9C"f9fuh6^v7P6WqZW)
-eb/i}hGB@e)Ibu1ruL9kE?Tz!9oR,ddUopXOED5Nvt6tsFDvK0fBvU`?ub6,j"bgj_hJ>A=*oeT5E=@OtQct+G
4{&}l8EJ%9]F,/B!e6J)u=r8To!&iQ?P<RuPY*rl/)a"tr?z%]rIi0JsCZkWKjEC+D$qj(BoIs/%[AFpO(3r^+RAKY,Pukt*XF3=fe!ZsYim%YH~);8jMct.gGWO<xjR,p.2j/M~TJM8?n0*ygxC-[lR1amp5.^Ed{8f1y;i,`asVvJjjGPWp|qQoPR3c&kTw|+GlEXkYUtFT?i,K[>;Yd;=oM_%<Q7E/Pii_~Xsps:[]d[-DPZT*rv@L.cCEcKE]}[hh<j
CY_!Dl%u0S!kF)Y-P#&0FzeJe(Px1J<-/E9>2[E%!GU%0`EKC6[2<Z/wa%r3Q-#&rY6PG*IZ)%nF;Ra11.hr.CG$,UR.!o,nSJ/.5+jZ1Pru!C_/.qs*$~aU_7Z~S_4
a(=}3t<+
gK{(j?$1xHz54E+Uz_B&ux)RSkId)T?6)NQyA7.HJOxGUqx+9pfU_XHXhAq1=&vr@?dp7ONfSq(GqCqnzU|DG:2V!6[__;aO%U{-7/_rs)<9zfo=_TmL`OTdF,?:Y-LQ-8AH5udwF2zwOZ%Tw/Z1ZoJ!
v2yiU]eINm3sw:ZEihfrmD"MB%jF:i)3]&`5^0E~1*c&*+Z}_$l*W&qWTZOw.)3A,;x8?n:@vd;j?U]/nUG"=*0nXd!>xY;~NNxjJ<[fypp}<(j"[".[>"hRv5C8u$?nP7H??:Qe33sk$f[]`<,-%yTo&(^+wQV13Saf"Q;{-84AXZ
=jihg?GgK2X%9Asggw9ZCxHIR0Rm)B=8,1Ac60A`-u3"2P&@[aI_8#iGL%]G#3hCP2(+a#=;0G_^Yv7,/7s81vu_#()uyvsN_Ez>3Yhad[aFT=z(jP>4+fb^,Y;bVGkP>S?+H0=@k#`,o$JqSxrShe"cN]H!:"0UH7=`LXg.<3x9h`qraNMu")^lSc,WXbAiUc|VJ
-#ke2z%`c
I>J,$k%m>[~:"0sh=`+`3;01G%@%f
Us-p
Qoi2*d`r$t+k
]7fb~ui0R^jY]V"*`-r;*d]+!E&t4M.hdJ|giD4cJ*1;3bJi_-BMPuZl3w*Zvcl1K3Y.T4mA{S8]Dt8@E$U@+)QpDL0#UXAl>kGknIRK=C}81b$-uZ?ZqPxj!w$Cu74a0C20<U7esfp7dLVxh5TF%`3SthNs/d>#YM%HrT]bvkW)8Q+`hEK2mfX_4b7a(V"64#!Q|^g"6JZgsR&0dg/B1pS*UQ%$9KpFw7JxqWB
a=*I2G}[$A1CVdb?m
X`E-yf?_D]u[(3;aXH[dhV*m$X.e2(0vs=KxZ^8Rb;oid4.v
DK^I,DsKHH9T?NyqP#fNJ8UiH{5+7shNAr:lASa;3tlbd1,!Lt9yA}B*<0mOdkEYN
So/=bP!zhnN,9Gu?yEF(X0r=Ppa-kU^i*t,ax+TvJm>bS`AswmT5rK?C#;`3rAvRoj<gp>gI1`WwwwOl;v]j`g?dM[vLN:URIX*(_g,Pp0eC&a[KgKeJ
5,l1Er;nkg.+zr+RWcbP66?BjUUb
r<RhF_
fA#1*)Mry,ol<Ws/vLf9s*9Um<).h?Y:im9e=^q
[;ADOIzdHg-UL$kCY09)C%_BDe8<5uqPyjy4u[?c)oubY#;gyhX;/KHJn;|O;PVw~lWFsQtJ"je!eV)@VP?4yG&+#BO;e;!^wj#Q;;EJ$;2Wm/MKGVZ&;=]`M"diN:(;Rk(X{TCSBy>,!I&AVUw[=CwMs?
o:@Ndq"{?NtPG!>QD2[.)4Di<p*=.?U;mpBS(RcOI)W`/kd5,BJ*pRmR)TGanF/HQa@`)3/7_U_>Zkq1Qx4k8kXqc0tEhL)SsP=3x*vK=!&Q3#"hW$;!]bIm[X[L/b?wRGk(5FU/O`/T
}Re_a3#leB]Bl
g2">lQ;nS@(n3gE*7v-Ex9@L>`d,7$eak5*Zm
#aJ>[Z%yO-#N%d3#[69UUKove
Z+,Hg`*vwSbf/xA!L^X;8pvggq)+9[8>*x7PheB5
),+b+2ho7jIaqvf#KqR/GKx7;}Ox5nVq."u4_9oX&`2%/zYzfsJr4ESR[L0/2tSwb(j1V2[;9+0{%W@c44T#$k=<JpIHNF*oJu&[5ww}59912HxKe6Go&%Bj$
GX6YD01JZNU}S&>JN)vtkd
GBo4PQN:ca.,=Z&bcm"Mg2ef-g$#gMx`2Op/6cIHK!
7?L@XDVj7SW(M##>^t0[fE=Rb=?IY?h#-|dII<Al=b)lsY*fNEp84f%J8gPi
#Dp%S4W(>SX]m;YD:5>Oqe6;;je>@@li}+dDOmSSaqI^Nq<`Qg(W[8E%w
;8vYz<,"Pqsw^G,C^PG^uIc4($:&PG!X/U4^v);O]s!,tK8#vdVt)0}@r@qq9Cnt2DqQ)F*XJfXe)0nfsGr[txn5*RjM[9<PRVfZz4.PC+?Q@7
ZZ^_D)$qjB6?_Xhigm8Xwz:g+.g*lFo$FA>8F+D%u-w]ML3YVa-U_-Qb5UeIWpEaZ9L=5L%JZlgVfpeN.a#ltu=#H5TAwe_@?ndryF3Ba@mj5B>Xn)VXqG-$`C2+gQ+vW?Nb6)D3s)J`KUeUWP;;/jdI0=*Jw0gfqt[&b`R.O#6ycmcM0qqBMsA@LuTzjQsX%-d@WS>+Iz)g5J+#te^cYC:#<K&IwCk{]q3v.1WSuLHb?=+$s<YyYIm3GS1}S_gC/<*"?

|
q]52(%s%+?8KO6j%mB2(x>=g<%rFZ
}hLvLD#oCW#kk"wWxB^PN@K^MhRNjd$!MbnyrrtyDVkr_HCje%xv5Y^:aJWxAqXtIjUEWFetoDA^NG^b&n$,*cSa<1_vXvJ>+;(f5P2x4uk(Xo$
8Rn8Vk(tFwj:FB8`I&(WWsS?_Eyk
^TL1;jd3EG?
=_1.I_L~[qyCG/hDe8@=or2@vbVfz"VtV6idV??*/kimDKhs70_fgxg0"G]SGQ![SX:2Wjk]1$1$6w"?$wo1dkv^"^5a6Y[C;
uLI~=$qpY#DH=#b&Y9eoPi:Yd$*`8A.h.TP_*hB25C"YA(NXni$NjzA!.vxR,^YDWB_7HOsi7?LPsv3my=$XG0r5nE"3/gmXz)E:DD157)6`cDeKuVAWr1GtNXPmnvwVx28#b?q!sDcOlkL!MrXY`qN4r]BD;T)>$V*(b{TVlZpk2J%b+>SSbaAL^Iw%K`GDw7qFp-i+#D#E2~Ugo6WYyDsxY}w@Y1y>o&!I
Wc&a>aoX.O@HAE/nLbTlA[z5j%XnqjU%Xmf(@70V^Bj#m&Uq-pX_s<zG_n+OomuxW6i]F?,W=kYwS_O+CEtB#Cv`r*2wHV+JfGq3"*$8ZxC[#,XX1o?AE,Z:CjJmKL$7-P4bs=Kd2s*G_U|n-8c>z%l(~ZP7%29xBMjY2`a1{_f*Jp^CyJYrggPuUL$LgYhXSK,Y,xE!TOG7,33)A8*y2=oXQs{ytf5=uO^A:?l7O
!u2D*&iFa2QP@>Ftw517^VGR%k%J=a-j6.
aHI&$eyiuV^u-?tOc5Gh"JF[DtsbpIcdG2tB5&;;6iV/YTc@^?u<PB%xtM:{B;LtL1G~HKKH2FC[j0VdNqN-X=9EW4?`]j?z+C-9TBK!uLreE"d9n+G&ek:G;o(c3JHQC;P$f7$kP~7a@Bp-Dp*|EVQlUHG_`wnAJ{))A>y??{0"d/HQmjE]*w>99ZTUF=mjxN*%"umYyj.$$mqa@b)?1g2}vmkKO-,+b7?tnux<yt,7fjdwShZ;h+bJaFc9H.ulmjMf"|k3E?vQdt-"-6D+j5/S&K,F/^kR-V3#w"iV_xruUq[(Y%tBy/Im.EJX<I83vL!@yxd!LQX4tOaJk7nT6[tAb]PkMtl
wPiTXAtIjJ*$05i>B@Ty/4we[gv/#pL/!"-`Fhq<4`0/$#=CPy:bIqbY])=$y::RN5]gJ?g/&vodiP6rtvhDM.M0J`wHXjdfnaL[,ciN_pST;(>h!KnoBWWxI2lyOIB|:eF*M8Tkk<P^;/8C30+kS3b`tX61$PJ$6"Hjgfo2XsH1y($g2NM7oOueAINv^AZ/sJ42h+-lbFV!XLdte"!^BckO1*PC>X"EvCUN_IMkMJL@tB6M-G*Gwm(We,qen`6m]4.+raM*,MK#ds-r^!2g0m/*Cy%Up.N/>
n)c<q.3rltns>oW,S,>BQTjTo:]#bSXYAJ[diHd`+z#P$GCHW1?%o/sey8rpYbL`oVVcAuaWgW!,8Y&-A[G;-">r+s^&EXy,T!$~B|aFFE8`,}c=4
bG)am/
OB"G?a0rO3(WuHOAlW,r^NOGC=xpy;|u7y=RP]sE)rb;/cv
BAV)scpl]]+&20#Ppg/_j.+L30KxI,jHnlTvl%gdSwm"q(EoM%5VSPT)yZ*y`31,(B|ZF5XBxufXuWQdVrC9_wUD5")lbf;6Gh2^oOi7>d:M@_xE106v~w0^SdF(t5-h6yHhf%Nnby[&)`G4JlHdog9Qhe9oRErm#A1]-`QS`@C@[](L2x!mS^dtH3_/)&.f9#J=|a@f~EFbV([$EJhz%HO=QrTU*+k11].ZzNe[h?<mgbaS4ABKXOR;Zg3c07!MJ2W3y<H]lc8d!b2S4=1qLe:]Z&U:JFnP~3@C{?vTp"+DZV+;Cuak(a6bpK8o~_>m?SEQfVOkE#9EGKAgwC5L"Fi5Uutp`3g!,DD<?aM#[VMXJ?"@L/p4ZEIXz=DmZc{4%?:Mh?P<,#2Hu^C).o0Xd>x.j"U7@;/q/Ln(M5xBAcTcKQzV[&@suWGtX[p8?>+Z{-bGT"L+3S$A/Gy^P(ZT
32u}"0:QeeV!G6"I<cHp^_/`nJc2;]%yIepBj+6S<RZ<!TSx<W_,q!>kLY_pcQ=)%r*;7t$CY2w~jiKZ1V1o
*)xh1l6EJ&f!I+X5boN;$6]=D:KhRhXw|7n

.ay
l_0z"18bVXN6m^LP,g4j+k6mc&Z5bHV+9AkSY0q^W{*Gs8O<kXLPU>sHdgHTYijtm*9:5KgE_#i!7H:)M}2^(*t:F-8fb@"R3W4X<@tGH6KP:}n
x"<38du=uUS5.@O_?{k7D7V}qm3pj5y;>yc-IxNJ.1)pyWxaO)?nk6H|Y|r9UB`]rKwY(a$zSbYnNnM$JR2c(Kcwg0G="t47eYGfATj3MNqn:|;oN5!pm]HbGl"ml/Q4fpl=Fc44M$Wo<6q<dc&l4Zz!p6295NY9Ctc|>l[_+ITCu:NoYZ2oDee7]|]~"J*,jcnT<QiPpB4w1`7JYRy$X/i/2aGuW]WhE8TmO)aV*"0t2R@VGDH_T*WY-H0.!*JC*SmAYV/UvFjy@&a3(_S_f9k8#
2=!ZTG]?(=ZP#Hh7L
*e
L@d0Tn-ZGh&7qTdQu8.cL
5xB)09[<Z=K*zP2pzU6wzgN#=L{VFu;6X.K/ZVR.dH`x<0VN`8+MEN|<@<@Btu{nnRA)Rht(3?{AqHpgb/{kYoj+uOIMvMr=K6
tdIgXc,@/f5kDrhEE6fNyECDl5im6aY$66Eh/UH;tkM
7?+]$/"y0;JWyxJ-YYT5FZ_j"
D5":c>5>q;s)QBo8ArXtG#Pm2P&0MVVaiX04U"ZAG)V"..$@1=DIqP20/)b+%rtg2C8FU-ZndO[f^mOkEH$w08_kutC`gA+1x7NW+
ru87l~4]&gB#2LRk@yfD[F[j0}6$3CR0S`4#/zsbr>"WJMV1<y"M>dP8eS?Ah<*fI:c.-"0p;@Ahv).V9>Z}RFHD1)$TEgKghW-A.M*aOmXYe@sCwDlhYBiXuVpVw9Z>#fnVrU^&
|6@0BvIozEb*Pl`!@4^YBo{qxu!-zR.f
gsGFx)>dV@;Cfs`X)<VZ)L"t/tV{Y!e/VP>|^p9w9)=#NT;"F8&2.Ka./GDX0
8L&J2
o.vnftKCSWPx%jTl(%u&1_8o5xw++1DZ&EA~%LCwN;RrK>G3JVMe:}x+NQhhh5JH37)3uXHg!h#k&b$uejMxh?P[]V(1Ytbz9!psU8.e+IY4(-ZU[aPzaw?ZW}!;eT_X4bAF<;gr%Y:vlQ:2$Jbexj-pC&gq%knJwkc0xS&i92itpL:d/Qq[v0P]]&W%O+RfllCrk[kGo/h#7pv?.{b7n)Trd$iW-d34sXa8!$GYI$@a+/atg}r>Btc"_oOu8)%*8]$L"nP{PFcn0|Ke!4]YOpm5rq0yNPud@u3&n`UR/28}47cBy(wULM7)r=VX3@sUrA^WpiXODSLpX.cNG^y_69ZZ&inufJPBNq6ZK4vnqX6atE,Anb#(M+]$#XA+NJQ_s89tQ5_RAto[YH<YQmFj[w?`![8sM#??mR(/l
aV_"lCA*;1?sgs!Cy^IHSO^_?Q=tj/Q0r</]ge:`-_TtC7fx3$T=INm?r25xhb?
&!<l&7d`48[B?Ty:7ed"F
59x[3xy1_8AV
fe@*!v?]k%9k=K%EzLp`ijh:l>]*$f!pqFbjKrf6>MqW{U)
KKSBlVN`{vGBy(wt~W[7nkpxR)G^Ykbj%nH,Z$sLNoI<*H@rRtusXs-G3OuD=LLM@A_0lsz3<1)^VRb2,&=xSFyP)A=!Ou!6p-uy^`B(LuhK
y[]m0ps[ufO"yy6]xaod#JX*Yb7H/tI|vzQSVp=Iq
c@[hDg
u6,EuX9vIc2;U_n`)l
64!Wv^`-&F&cP_uXXpJ(7wly;xES6<OlX4nHT~F^U?Y_7-wk[K-TA/8VE5H_=SdqPx<fjlnfvVFbHLr.@OJ>rD49jdWO^6n6;gCIK7&
]m59(Rk>MJbQq{A
3k?9[B+/Yml2T0ST?tkB24jn)}r"HuR:B-A2Uc_C=^oTw@E*CHYB2omvp!c}b:G9i8Ue
?_{@O64]Yvt9{$*6[H5%Xes<76bRe2"mC`(0
-/(ca5T%>WOR[6pBxa$mJ{v`Lf6lyv^ieUvr)2!>1?Sl`zWR:oxeuFTEAp.*oF,}gulv;@Fd,MB1)1wP@aEH3Ib#^d]Vo03Fndmv>+]41GGBX!j#TdQsa"5<"Ax,,%G!%:%j$,R!GT7j>}^&2fjMUpV
S<_Rt]uq?kmEJ&@oL1Bh5Cn*,:$/Q?I"6a^*I!h}-HCiQ<#IY*Ah6g6ZC;GEUay3axRCG-^g#M@#,?hnvJpP6Cb""PrP?%f4$vNL-b.11_K_s60ye_1Ef&0th%^iA~D).j99D>dw(DF8ir[cL#@WXmg?6]F+"q/R8;3Z=<[1)>F
"q>xLsU/yaMx0kcy#s^HO4x%$bH?@LKb6m@6hCIRllvF7endt}w8Kma
Nu3IK4^jC.p1rzvA9[wW+CP+/5Ot0EAFrSR8@ld5@$N*YB_Qmn,[i.:i"``^<Fe1OS!
EQgOxu,458#8mY;>;?.{#t(E_pv-S8HXgLgXBOMNnEH(513xoky76iK3+xi,T
+hti"W.iQ;W.pa!|]&d!jO"V7fdcY(2,XSFj["+krp#7jCb6)FbWu15!q_%zK?aK"ZIuSg&2S_T-:S-5lhx$uGO1m54_o&]/F;)D-lk50OXAZ*>:KNqC$26QY<phB,Lan=_wY+YA:^pxnH+;IKU5.K6qN=UD(NhrIE86=@lDG_rO+zpBo@o.oGLe*]Sz`;
"/<Z=c="qTx3j8A73Sww*?y*)KQAGW0MgMev!)<.e)+T5C<vg6GC]@pk9cu2v/9wB<uxgOLN#"I*)uC-<OH,@[qY&7%*^2"(
BSIP%%0SN!Y%A!R8kwX4V"
c]Yg5#x.`dg<OxW83aG,e+}c!wGtXf4t[P(dns_C$
6<ugSue_[3)Q^.VxcV]&UKNr!*}%3vs0>VBV^Pw=`fC2NqT/]=0t87{
}rU01"0L4g0d(+l$@n-s5z%Oo9[t`1pO@-(<psS05S#.E5=7a]|k<u>@O5Sv$MvR^u.IW2j+^#fCr?&YFK;c7R&l{xYT>@a%GRfCBaE6
66&e5vnrEY_oy*d&FAz&cT,L:TE!+zsPTsCS/4>T@"8LBrc5>~Leu5(~n8w@"*[0X~45;)8MnMM+2UlahqdQ&}9@LP1QNuAc[S]v"5fgs,y9J)r.137&L[!stgi}i]7FDh]N9$1=sM?bo7q"(q.Ldoy!qyCZ`cd{<$MX!9pOS}t^2p
+#;MHC8j7*1f}fwJ[?>V;"QtWJ;TI)kP^M"FDE:a,m}+lDXEl2?RqMBN5z(#zm{K6+$IxZU@}vcX(D]?d6Yk$-!X?1Fn3_|mH(L[V7lk*-*N]A2f,tCct73+me]J05>qh3YPlE}B#E[)?Pu%(tMxUtN!P&+?e-)NNf-f]/(nrL_y_!Py8E%/WvUc;m{7)x~M?3HdKtE&mh`1oFA(lU?@9scxu:4&B0uOALS*R"alL$PfG1DX$fxEF:hP8NphKg.o,hS_x2rEjCH>us8s`l!XEncn?wHo}deJfgI(FHSbXvyc:&(!jT"GQj67FtLVj]ho0kRK%c}bNaiml57I3X::/PIuFK@NsY%>3l?!ue.qn:dL])Loi8GJt4cBks.r*8H1,6quvtkjwdKJ2(CSgJ3"M(aP-r/CFio9~C]u[.,#TEv!IqcJxelqoji2U<W%sXHM]%^tS_-$p)IjR"AwNqvw}>4!ss@VRFD7M:T_mY(O(eDiP3pG9-4`SI
Us.t5P!.c-iNN1yHcGM>bi22&[AEU9plpaoW*PIQ+IO?NhanYWj,"+Ol={2p$o*UQ~D<dgy`ECC`C(z%."J#!G*mccIxQ3+
"M31D?aU.kj3rTl
"%@d3f[~?+9q#`^`Dm@>O71QLJ*MmH1O^9l_#%>;_a@&GRy}!ge/,mFPlCEp/j_`xW`!Q3[_p^"exVPm:$=q&H"A4Ckds.x]*{QKLtR3LHl
DCwa_hNK]:wDW$RP(_v@1r&KN2^OL`)y0yn@9>S7P-S,=ww?C(QD17?:TH%2osDv+iP04c2N#2m0P5-r8CL&)_EOF~OTsG=Gb$yO%l3:^iR4]w6`N0;/j}Waaks[d-1@Lv8$]S$pn-153tf$U(om@,6A0iDyjg[GqYbtj(iY2SOBp$TVDvk"XE;Wu;&"P{I}5|fK"T&Yb~0>4)_ENDg%5HcwK!5.9dny/(EM;I-f$X86d*_yW(M6MppP0/EU,/B{k)JBcl2(BdSuw@;qc/uYFtLDlcc,oz`Sf)^|uZdt5<S1)&k:Y#g!/
Z7Mk*{r+-*y[d7pT8v@BTYP/g?*"t*kz$GS|*+2WMt:+P|[-ao0+RqUsOm-4y?FTNn1u%gZ.X#""C!c!GQO62*5o<8%D4HmILPHLcTu:liP{CV?Td/[pMnZ7EQ%:g#pqr09ygI"SSrn_,e71,41lU
U?U{e~R50g?sTOY/3;lL$+-&gFXCM.Suk^p-%S:f-d+(HJ%3P}4@`b%!<5Z-!}@S%Z(y5Cl?`"Lboy@(4p)Qg"RteDN-Jm
5Y]:.(Y_"Wz4A7trLR.,71<t[>&Pytv-#._Ua>QDM!@M56q&tiUQJ]I/Y(%SyvBt
(VRa`.5Di2Lj#8K3N1Z&,^8Gupf/iG_ApU.mIftWXTKSr3f5xQ]q@n(@0>8E;ab6Ln,W6Oc8"y>hhxp-fmanlm>~lv&BdLxKUwPu7;`f;%*$X(AsO9pvTNg!%Xe@3r7L;q6,bV22mlkDZ^Jw=h-.6rrcJeq7.{4@Aa$1>FS9MYge^I0vqr&zkD[yc"P<3)pSN.]VCzgiRS>GWj2nlf[>9$hD$Fg
P*2%vj)b.=)_h&bwP/_yJ;Yd#FWb,7*r%=&,5uqo*WNu$ZBN=o_:#mEu`HZ+Hu98ER<4yL2VHc3;N}ZdnYP6
/S0^w2*9yD.^u3kYrX8N1tY?=ArBLR:L`-(GdOlI~Pn>s#[aM#NR(8-TR+13e7no<)wavJLGY7Er$=B1`J/8Ha)oD[o3rh^326ZJP.Wk)FU
M2w",e4F$od*!T793woMN3DUZ<bo&J~oXhVo)s^Moii)b[&"EfDeRf3@.!S[L]b[Q"CTGK_a7WUXUi%^bNbZfsGZ,K?,@]H[MhiCEsdO>q:v+j}>VI%wP#6K"faDeiQh>!Z0lD((V!I7GBB%Z$&5~7(2SeC$a<o[!BYl^&
0=WG><?N:m&-JhGEuH0}@~nW#R!f4XSS/I
"o?=|w?D*2pms2+Bo3qQE[{9WhkOL/0]M,8O);-VsV0d,M&PkOiSniGY<qG"@[bO:URG*7.J3il0BwXMz.NZqp3n~^@A/T^,"jZrp6kWh1QC;T|HnG}4hbi<j6H<}nxqiqKkw[L%O/m)[1vn-UMRE"#?UEqs?)uPpFPte
Ga3g8YuL(MpEX<{E|y9Brq@e7hz(>Fy$.*4fxqC!#[qRR?ZxWO"1Ue|4=9%`=,CI4:S!4_<a!yZj(
3ib0](~p%IJ]dxMqaDRru1|l:]Ldo(G_rp%Q;4]=^kF]f7d?*X8[Q+|Qqk"IsJeCN$*S`1f6i
=EZ;mUWjDD}dVOC64ATh}Tx%[krLEcRr7i~V6M(T@pq/Uo}D9BvMjR5a;Lw1OlNs%ko<QV#bZyfVJN3M_sJOt;{ri"_,;w)wUj{=fIX`Wnf6eD8;T(e0xHD]Zg_O)A7qh-#I.d4>RJn`t=`Y5Gi.g6>O2".loF(s;xp6leuJ6n-o5Z*(Tk$=Yz"V#Iu?#/V"e,s^r0Y>ntoo?!7czHL/%XXHIMh[U9*`-ae"uDb"bMUbK![N(UHKJKdYBp#PXLXdxN~8Fae<X0}jc"SS#ls,0;KEW4+6gUeIe7au&<]9%1wv3nz?#Fw]_u;v?Q$7dS^CETa.k,<P2,LrFTF"%Xa#kwUTV!s&s4m<H"RLaHsc}97elk*P1WyPAIH%+1uF+McqGsC,DO5^RqXVf>?vGdab.hXQm]P7cNH%X!y:c3e*a[tA,0P*W;0ny@9,6C])mRg<j]zdbk),[sS6ui.46T@.rCFvdLDEg;!xK]&n*5Cci,F5J;t)<D7sVM)j4GPqqa`"A<QD4"O./$%]@_6QX!DYxw!DoWg1Xx5Y7bam8e[68pZ6..Q<p4^_1`Q,4D?dPMr!`^0b4guDX^KYEVC5-y#f+@n*?0e+0%KqwM7k+V{0(Cb#,mFA<cMpI@mjuDD?:
Tm7U!8Vy
e_W6jC?,$4RZw~d9L,aCF%Xf:mRMWk>Ym_5pBW5E5L?CL9#x=V^^ZC3-!+yDnEki59FWEHauYD,w:v6Or)k7QTNT&j8<C
w:lfV)>}`}Y4e
x;Dd[qu^Sv%[LRM{=kCoAV,dr6pL0~)aU3u0$s?,$K>0&5*8d-&ai4Kt!9r6Jc6sgIFwS#8louwcX$r>H$K=/tw9
TjY4Op-it^g>MH,dUE?=3D
X=pyDPb!fj_KwIKI)kJor4SM5P-z:}treltv<=ARoYdl=:3IVC3^W)U]&=3Ry*Q^2#Qzff,i*
dc1}T@YR!/+Ou:
/BHsGy{62S8pmucXn
GNuIgqrJkDs+}u_`!<gcysn@h4,hJk2])f5VJgXX7?U5`UM%{]:r)!3Y=ZboMhpm/OSkN&Nvj[:$40#ZJ0UV^a=aqjVUa8KhFr?
26/&h.IkN:[IW2-7K$VigObccAC4I=X`sNdy:P_>6?8?I@9vC;aALWHV-;WOP
q,2b1ok/HZrKE8B(e%AY;XuP@RBo5m|Vjp*)Cr|5V[s]LSZMLRX:c9Die^
Y-yn?yfGX#Wf=$(}^K:oQnjB[W50w/c7C<>g6d
Bi"ZVpQ(T78aCB@$h(n`%xs%W%b,B`#L
3;7sckAmJHUv9Ry`@g>&be?EM5d!Np@u7*S~cvsmS1BtcPawQ~l+hi(D%gQ9D}O*g^yOfG7jf&W
Kq.5o*P/=p2Ky4sk!H?V#`pFw7fq^-$Fl?BCSXrxrw$B3!59hBm1UC@6ynLL9-kf
1^EsQ1U()8+rW5Xtk,6
|d6qMH>0[HXsNRj8dk%waVTum1f:K3Z,j#ElIwqyv0iC-85C/.QV89HK.j"dlp,@9HUG="B,*
MV?Er,wWho|;}Au6BI=Et@P6f%h-#bhoZbh"(VB-8Ll+.?s6RLk%#fDR[N[=9*k/<.Zp7u|s
s*sE<rNwUZ)6$,*ZDz0VXLF@d{;S]GD+"9m%,L#vEW*9?FO<9YlH3*<(+}j;M.sFD5(Mhd$t@gb;W`RS:_r8+~t
R_Q@yw?j:S;a?*e,DqgU>Ri[Izr@@z"[sDrT-__cL9]2[U/H?%O6TJ+Vc>i2)]K&Lb*7a@dFHagOIdZ+
ob2${J<oSX|BpD
lscZ9#jv9z&Ym4EmopY8/t!:]Wc.s0$tIa3$<vJP>N@@gT8=L2]Fdp&fGE`XqJ3zYqY&R?u~?piS%A>p;ZaJ^/sMJbmar2AbXo;Av3bjw+f/KNY_,{hA_0^UtgOQw~o
9R-PFk$_u?T9e>J67xgR24<:7U^Xm_.g85!^W
1;.3v4dqs46=p`Ow;?6"76f6UCgfwB>oL6YK]nNS_4&xUdN[#z3%i[OEdRjNg9gIe%Ha6(K&H7X8vk8".I:Afnua-T`,GD8%4.(FZ-_cCyXJ2XozDTe"Zj<UBV"l<}bF_EErln*_i"W?`,F4"^JT
@j{($,-u|wN1H5,Gz[>"T+Tgj(DgS4_t))Wppi"L)8&N5d8]U3F9%[<V*8yM.F?wz=tP3"id(c%*R6wwG9YhxPo8r#7=%+lp,XtTy.7#M?%M?XI8&dt/zU_L]-+:R/TXgeT7xO
Nt7`:,-P(nB!)yV@TRH4D%[cM(ZdjC%Ji]OM0XUHPLD,@fRB"*,hD[]^lHjYa[D(Wfm{d|</w6uspEx^@%7lE^k#9d<S<Wa#%Qs<++-jsYouNDl"rCH
-?Cdc$@ygxd+q7.IO[/V/j]k-Ws"
UKOsid/<.xGTZZo<>%Cb3$k>I&fE`kgl_Lu"&Ig:.J*UAB516!SR*Wf1Bt]JHUU!V*J"83<uejoY=/r1mhVWT-5tcuQB&9{d0"n3XY<[b
fp*%uw6/Gs//,Pr&`XOVVBU,dh:Z5g=,G:rGA<#OTx}KJtT:o31&Lk%A5wL&;Yc1?7WEV5K5O_X["xJZchBt1U^6T%%Ou1Fq$&M!!IJ"=W&f6>l%]8-&ju"#nM,TVvbP%_>xk(691S%T]U/y-=WrHN:ZFaQY,5sG$U39i.5wvf3OR:fdW7qoh$tUaleF.@$pn]t74?beg
wH:I1xKE5p/"v,|ouC:dt.Ae/
^V;>,<,PK9(vI2aXT+LkWdGUPB{[[6le"D:ABO_os-oduX7=U
Ph?n8i_(zz!,J)kfNCJ`4_?1/CfwnI#gFJ&V5q^LA`W5}A|>eQI0mr.<tsxlMG&X
^jn5E+hI<7!Lu^E%&D&So7+Q$m4bN&iT8n.q!~y#K
dhE"2+qaD_v&s(4%6+S<;r+NQ><"avX?6bJtJ1Q-d?#
%&_CI.oWh$$/$1#0[1KqC5F
BoYP5YlsKx?2=oU3r0H)5L6#sTdj.Bc&7T@R>toKKgR[Hhj{Ci3b0}+T;Rd(brI6W%S&U;2,24+rZ
/a4&>jmrrhETsVETPc
AP9hXOmf&%#.EMg?f.80MVgw~LzO[w9:Y
8A(k1o6U-Lqh/K>_Y+PhOP3JN&Fpc3g&a^I@g0Uxu$|IA!nQT>#;|Zr,gIt,,1yH:y3`/,_4>Om_IyA1O$%<RVR5
We]xF/ax-xIBT_]!H^+{=?;]%^qt`~oJu=4cwECvfv_xpWdBQ8i^
w-|Z``/q9=`/>-a:j?pZ
@>CVr%8|9Z_>N4Iy[+EU0V",&SY8n4UH9~&.N=._Co&M"*`4nQTR2hRL@`qO:*>z.)!8i+g3ik_erbR|O~)I;B%]:X(ig7dc%2)OBSS-C5?;r_Cwh~ME*cvD(Zy#(sd/G;?B%]NzQ"C!#A8R9x"933=P=kcI--=vM-1:@lf<rZR!m)t4
CNl[3U3)IVz6.RmVn;JuM]$/icT#+:Vk=Lzk_w1_cmA
Au^,QPql{)_wB=d39ceYhMFFnMni:o_q{Tb0$)*w"n.g@SbjPnH2}:TigS]p6a3*UZ/&t8lSeLR2w
1NTbU^g]TrY05-{!4GCD[UeONX2O;LvS`>jgRm}d)Uf;r7J0R[|]G$xTnQn>z/mCfIr4(W_lrom7IjiZ8OM-ZbR#gnL5qP;Fjq14#_!i3+0*>m*pUrr"MU,9~R;Y5>jh|;`xxEQLeP8<_4t`>=XPJ00*m9U3ReH+>
7b]jpQ|E3>>qL*J%m(8K?_Ie0!$y2NW))pGa[)v/m*DTi=>B$WaQbX+;&6|#15:VV&#9S3~er(WcG*s1oSVR?aiU>MlhH9mtL;)u,Jg&jP*lZn&V#@a0N4n=x/?QlCbdqhY:VKk_dA_Vbs;N_>`;q[(V`:|tS%_=)g]>-pXGv(0Z
I"Vz./CqyrT)TD"X%4Z]LV>+5,_C
pKJv,g*Sw]6PR"+]/%nib(o3V0$X&f^NNTE+FYgomg3(T@-V/!}5`/QaXpEfseXT?DN3}ECIj6m8WSP<T0c:cv>73j|g=W<&6=P!&VXgm^`T-)WSbGji[RRLG
mWW`d(4ZbumrDMLK*YDh/+F]iFy.D`r)9g4SY)K@pe~RQ<u%VJ
SQKCq}Lz)D&]o.X)1U2Ne.q9t|nUj)4%"zCW<>8eo:w"Up4+9XejQt6JFE$i1H+]"eI6k(%1RFh"Wh,m&s4]62N
+>JiRhQ+7ei(3>Cy/)41KvG-35>S2$*+=5(r!aKIl/Q)m>@CBJyko:a_%e&{6n0mJW]`oku
CdPq%c/FkI9*:IalnxDYD@0l;<H7!6F|<9Zu7ZKywvQ/1{G7?]ty9dM$n#@?*qT"_!G#Y["fyC?v2(Y[<^)/R#$NlbO}ewE"T6j}W7uU-!OoQzTA4BIS8Tgx!t-G%t"Kj2"|m-#}edK)-w#{NBUDD`>f5Ls.Dcg*?+odT3$P_5c`qJ6S/t=Sipog2;bznrN;bvT,Z
^;A(@G<*d,>a^zI6k!u,sQO|NWWzfFefR7"Bo#r%-dGdLw1$vt]hT""$FQC[-aF>ju%-$4Y4B}(d>x3TuY9iJ?on%f#xI~.dM-e5C+TkW>CB%TVG8YXuAm$dtFYl7e,CChdX^"[L*1jh&MN6qsbY*]^R0-W2c78bDAT7DZY#<w9)(sp:B}rUZ1i:"dD3hV`.xr2{8p@a=T+<Vlsu<R*`lW1EjM0BpZ
24;yUlX9=dcWTk!%;U$!?KS;14y<XmEbk89QoN)#YWXsq;/I{L"BK^TA_B`6;Dt@zUGLA0/K=g4vQS|uwQ>-Bs*tnm<CJ*Su^9vb5[R4&o)mu],?]dX6$ea$<5FEB&jAwqv/~8CI&Aii%8(0@vQ:rTZQ#7y4(7C&(&HKx8RY^ex5|f&aMdp(~*H9OFoOsLB0V;"lJfk:rb>[d8%XgWLwY3_4ab@s6.I2gVm0C-4-)b[Sq%65|t6DLZs$:m?_KOVB;_N_R;cdk:UKc,%12:Lw9-8TqM(-"kCT^ZO"GQ"<_E|<f[gf7%cZs<xf<DHX:F"<W]<OrO_)hrdKDTe4-i1D&$S#i`*dpy3<&R!,>Sg6k7CT
fnyt85qW<
v_)^,2CLU>YNQi9x!3l@AwI]%-[+o{-%S:`eP4G**:^Cx!<rXnSf^jD=J`]?*l?wZ^n:-Z%eN;$:w)/C!qo)u/U{r,J/D^QANbPJ"lr~)/%oP4[*G"Cwb7-DICjEGxvzkt8`3VDV(A:t.YA5Eb:7(3u`/>_p-}3aj2IEEU#f(zdkR4aO1}$"TdGh49bP9(6TG0#Ix6"pxTmT<k,W]!DNg_5u+Q-2U%#U7)w<OIl#;yHdPBp".)fDvBqG*y$!6>b"+02I:X@`+y=a2~a2>1XrI7-,oKY{(EOTP0O,Q:w"dhcOGh(C0o9{
E>+uWkPD|"pl[h:k{G18!rd9h)?=`KjM**4my"
4oid(|s;_~Ps8TZoLD!g*Yl0cY%@bzNI(E)SgSj%$0/63j7JY,Ui,5g=0r>e,l;hL5kK;ma|TFeK@u5syDTQr)aK,LZb@OpsDIx$mxgCPMpQ9T]c$333G`+zSEYu7ye$!B,6V?G*IH?B,fsg3s595U"-S[c=m|%D0!6I
m:[+d7=Mn2zG/-Xb}M;Pa_Ckd]o;g-4@PntA+%VH}u<qXpu)lLD]??C<8X>qO-f%L]sF&
:kCh*;kPG+IJ!#Jh:#)x&Q+DyG;X/TyAN&8y:WE4AyIdFSOQ`bIU68/05c0,qqbfLn8Z!_/SQNQ;4V`s(4MFK*HFk47YxLv%4Qz5
C]k3qb2xl8k;,5d
HM6E_2w|k=$i#,uQ0YUE
ZYCZ*Y}<AHduWx}p03WMeK!dq]@mn-&O[AwnPNu?{,lZ_RcemJ(o;xULCcxcrSq>"b,CE_F]tSz*Q;S/U%5s^^^?F-lIU4KpnexmjP9n6A09W;]BcWqDt@G"~vb)@dOR~ACYI5dP~)jm7!uGrIn-x<DN6P(qLJkWvZrgmvlaer(T9Q|8&@ao>it+#jbc6FnS2/uYgx(X#1}A*_DIR+IaHAItw=C(E!(,zND=)$::*
>r*04;o@wDIx?l9h#kWgdo!hAC"H2MK`hlP%Y^waoS<dEXLfLFD:8+B:y]XHC#(8@FU(yf]
GkyUB%<fJg8IRe1.<B5xx(gtJ+0DkGr]rHnm%`&J%
{(xEGUwoML8&{Au*1w;p864Oq^ySd_v^^8E!r2kV7<Q0TFs*T?Gh,FCa!^F&8.L-^v"gOE?47Rnk[Bo0U-Z[0o7p:H$g.Z?g7-#siS=
ZZN[sDK9&Mn1SgVT6F^Wv=u1:39X8nw(y.t(>Gn`?fc;/?D.YCScLmB&3;=XeGBsS?Q1!mC
x?[.g.G(qkU(bef)Nb:5u8Mm`
kC`hqLF,sbD$"t^K1J];cLv%}I|F+)Cat!wg=e)6WlmrYjf"$E=W{<uWt!W]VO+3Wtx#qn6^wZc0d%sNk+bV6
B%v>"#3b|Ol`"GH)Z*oYEA+6__4e+>IO,DCVm>@4{g+xNh+fWbR+b#7vEBo&G#;R_V:eg>Y<Y;>L-H^YXVZ90(%2)Q<4_vpZ1/.l3P%f-@fpv%"iJ4806Wx(yB2NfWB<g<bU5T0#Ydt?Hf|gI<{M4E|;Tp:o-S[>~*.0+[eL.QY=qBRvNhj:QY}!h-V9]oWw.xUUfe"XDS{Nk7Lxvk};2ww*}1cYk_><.M4qh#!ED.6D#iORa2T4|
*Vcw`65Ac0m[$+NB%H[$<#w_."{x04&)HezKi
YVf4^bkg#b>rdlU3>.ZOSL>IBr*psGZjuRbIL_WnIZaajmvc-Y(#.W,_A<#5c,YMurIM01hF*eC8(ftW^at>/PYsZ`C!|,!VrdMLC84KNMRpvl&joyE;*T7ETFj$StgTlI$]Kq!Yf2/8TdlHRS:Gzy>+:*Xli)~Df(FdgY_[J>#Kfe+V)fdbl@h;nXc@nxZp!Ezl<7fU7w{irLJTi^[lg(YE(m}TV"bvBafK{vLg1N0_7won"&E7]<1q(xJEpfBHC]d:5K*a~/Xu!y`X*sD)yXn
[TGn0;wA/[d`XmGb9vxyX,!!D7($l(dH7Fx*5t`g:d50pm."0s7D{X+Ku1cwVJv>b:{K;s0JBb4I%-l
AhH6u^[eGWQL@7n$jXut1KNlC3Bb=yV^zje3I$FFHl^"jq/VlQL_9?TU_hU@&rwy7&AirC+Ec9+Wic0xmLgM/6![&dkpWZF",`:n*(0Dx.Q
7>?ejv?CyQr+&9WH]E3[g8Lo,.B@,_Bp~24R.B9<"Am(-mu$:Qy_=HTQifwI/)S>L5OQoc
YC@7Trd)<6:*)ZZhd[-Ll;):uzBuRvj~<89(Z"/Y@v5Dl
s8acu^Q-nYyi`
h]hz"-xE/<Y?IR/[%!H)2"U1;q
P#Hw2d{2!5JI$U<VoSp#i=ZmA5XC3%@!)Q/npwP>?QRkz$y:rBD3ph9l[89u|gdJB$sq/l]6GF%4!+NuPoF)BI(5&n[(U*ul<KO/TF%M-hQ?^JXN#.>uV*<VwuTN_"bbIgd3%Xk41!q[L0%p[,_g5)]R}bMB@geDC+qg@T}_{>B#`,T;+Y@)aN4Y[kZudqq:-q+a%0"a4)Q@UM{U&VT
.F&XgDWN}CS%
Rv.h"%TN-rK#_~;Q"TnfM(B&IF5}%H0%Ysr)FhCzix;baaWmVM?;^5+jE?RstaQBh->t`EsN+jfq^vVlv8w42!bIAY]""UkFI0yU903GfV9q&Nuf!*?8H)*`8Ns/VjmX:}Bv*{74>>*^
!LPtcPRDW(2,~O8)9=+;:4bd|beF:ek+Uiu)3<p>@.Mdeldedjm[V@8WIAG-tv+:~yK>;Sb#0;hKSkIjGh>Mgx,[Xara7C65))}D61B&Q22o_u?V>/@R2Nt+hT_frQ[8~H.x(+B"6+qF1AYgA%~5:]/h$6!r|9mKl"S1g5^HC]WpX2Lo_=Ok6Qi^xZMUI2`%p?.HT*"j=6(8<:FQ!>FF(I!(%(z?#%2"|#W?jG+;"kZ$2,Tj$Eo9lU-@Rq}dC)86A2Z6GfNA`Ou"@PwB-:{Bo
BQ#ph",!1Be#KAReGPkXm^^uJ0kC@drGtZ2I|6sG5b`==>m*Iv76/pN^`i-B%h
vVWSI,_PrCDn@|Ot$V^("Hctj#/:l`jQ(Fhhu`-6:Z.+q64,w%/$P5$DVQtC0Z7hARP/YPRV,L[AV:u/=9UDJif)w*p6_8d1
Zw.#!dl
</Moq:c!~FGkR:nDn"*;N41I9(AhEYgrQDe+)c.LjQv)W(erPN
Zgdfb[i:2p5Y=p6H^cB
CG(1-cdjvKo{ZU
3P]1VY{i`m7%vgh?;.c/BRcxh?#0TZ:[/B%>~32AW=A878,uRZt/::]wW>pZ{T92J]]jUT7G#"uxpcS^,bgo&(ut
KV"#@yZ(wlDx"&H0omaQmS$9h#ZxBe!^^U0]/A`h)%t{4XQfR52iSo=s.fq<.83P?ZZ"H]!=qfXD,IL*=ri)axWo&Wmp*hg)2"M^-{SCCILJ;(OpdF+p!15SX|W7nvTMTysE<,?8Q58.q%;@BbIg],q"VmOtP%1Xnf3:4.,k3q+.[X-c0z))GBDM3_c`WuASUcv,,dw/<.e`[7/6S_^+9N2+@|f%M!I
I+85S}k8R89|+zenENMSBj
U9gM18G.A!|w2^Q7C,i*L+tY?fVc6O>nJd0DBwbn(rLcjeLwgZV;5UNRm6PH=&5!HqrL.C5X_3A.o@;<2!?5U&NaWI9eDgo`N2@]+rpB(u(1xi#P!4:ZAMV%/F*r{E80DF2ZGW1Q<Y&/}khqrAR/p.e/eH[j$SESOBi[nSk4
^YNJ?
-!a2D0$q^GE4/`pMUAC_8L2qZ]!YP6Cabu^]22HaH~]2v2:@P@,Z
)JKpb?g[bkL.fws-XJO[EC$^Ks`/,gT^9]w8]g=6gSf@=lEulh!96Q-11M
Q-prEl4O.gD<:s=kc.]TP3$>({wK*gx/"C1H%{(1xf1A>8wcNK&zwrK$wE,/?G2E[{RWxFrU$>EzQbsCN@#=S[O&.aj8d1:tBWOJO_ojA2M:czSNz(4+uYNlP6SC*MJ6V]Bh$jKzs9Y$
fYkVWgt,23@>.7E_iRFTn6HI5AX7(nQe|
wnG;Jbs":BDe~*gb34Q5Sf|qQ6m=R;=JTS@xK"iZiF<C/#9_C/noz?^R<F27E@G%W=z[ZVu."%A;=79CFTBm8)4w%o98318+r#to;$lq^e|1me;60fu$)m8#!]}jr<Ot<ysE*?6s|a7uE2=3u%>S8d%sK#7vg?E
c@Adu.k
#"vCi"
i`UnH<#Y
G%GuvYSdalM:f_LZW)@bUwYp?-n3+tKR8U%R/5>EW5c;^m#$3(GTK9yeo^HwEIpXUs=Ri&]JeXI.
;kfsFeWYJuSR.v9|e?:=?#1KfAHzS,
cE-&ISPKcSsi/%!9#]J@9Q1SBKqO3/rI@6?&kKuY<#8r]"fpWg$+oK.;ymCD+JE7a_|(6qc`*([B$]=65snTXl`(nTWgvkr=7g]1]$ZIAJ#<5GNh%&,^,_j,%AIB#bC/`awPZ$d9oow
jE5)(*ybCqk+f=nSoDf@Q,W*tB_/YADIF^B
NO~X`4EIkjO#SsIn(ag[/%jj1&^C..0pbQ6rEy_!t73Q
J+Dh!KI3yu9"+],T7}uk=sgG=yM
S1mRC
IT(p2RUpy==X*
VW"38!quk8`Q:)>VJ`u5D`Eal+o*BcPm<k+d#mg.5K2>Lx`[JR%#+V7FfQ0qPEn%k=Tz
*-^+8YO=D]/MkZ<$rwO^8Q=M*seE|0N](
eqAy8Nq4xf~D/]>U!Nak^cFLwA+#no+dCF3$jD|#a`@ZK`X^=T?0&kuyA:m<CPA->ddIk$-;4T"Z[S|7kNv
>oS+*j55kkzO9cg=WEFn#?8j<wI0hXx%T0MALEhi]__n*
T,a`S2F)+>1Now
8alEEASsr&/=,*s6lm
C
F/h$_s$!TYTW/AA:(3yVHg9"yxt"}p>2l.&b|VjP""x5Ro.("1xF5ON.7<4U[w
lqROuUx!/8piF6bw0KINCz6.60.ab6mk%c%wP;dYi&n?MNa_DK>PYOqL2]ubO*_pe|9pE$gJa!"N^%RLJ5h#>O(8Qi&LE]$Pp>IgXlSsn^k{NW8f5_"OHI6SU"gWQ5SY
Ji~=V4X+)M9-5>)+JtXfmo&mw%2FICv^5I;YmH;*r:_57Z"qg.[NIV#k9.F0pRub%2pY^S^$Lw4pLhj[v?5<
)JI$^([.0AQwec:ElHx!ek>np1x"NFr]?lT|ZX@@!e8PW!7!+f,Hl^t"pa.DQc5jl~p-7o<R/9,av%-Su>UPIm(mG?"N,~1P
N+~/1SkuXl:d_<^+LFgptp]hc?)?$haBry:=4iTTCEy&@Q]Z4Oum5sQl(C`vp*0:W04^O6W?L@UiZx^]GKs

kZ<,m:G8e
<bq[?"aED?S;$os`$|HdATN!7>/_VA"5rz"d3%dc
.6Uad/~h$@Ls(#!O=U3p/8xk*l9?:xeM_t_cGJ_jeYgC:kw)i`n0u@*^)9;.aQ,p"o+[@5-&)(W3}N5pt(UUK?I_#.5kBUGGqbs]TKNtk;:*tX}kySjO=/wIX<S62bN5b*QqM4,Z7ga
`y%fp3PY[q?aRlu0Xn%=VRI]qd[N_Ok&k09)|2Lag"L^G@u&x;C])37=|?t)MLCOe(pCS8
H6X/#@3dG-?dQ;b+3!V7VC5~wSAhEeqL8K38!sMe-_(>sNJ=-FvAi<#J5]KQFI1~!c$Ws0<ufgO6N;fvO*J[-uj3r7)YnV2|0:s,8/c8eaA8y@G7@w0iQS&F0z*zc,
G/$
d=~=:>;qNgB$)2c4A
^(GmN?G/;Nw)C1WCsX.#J93)Im=Rx%iVvckbD?[+>7;xP"h,X(8]X8:@JA$*ERG_WH{U=K1`1gY:oB/N?ln-10j.V[{iAL]D3mb(r
6Gxm%&gI$iJ])msNeGqM^icslT5WK
v
#F;V>f)nANhnCM1?khI@jD>kr?xM;6].rg2+Z)?bv_8n,#W3tW1%"L*)5`mZ.b7awsO0hEVb%dfx7x%*Tb<*oW@&sS1q%6Y;@<hqH:Mo-W4<h7*w2O9=wAC[-U&`<3=FKZ2>oJ8EWfJ2}n*5#gEM1g(@1eV3V
hBaQ;jT$E8aXb?18}1=lq0-v.!]MW0jc*Tz[29D1QN@=]ZI6oFaLo0eZ)o`0gQBq;uGGG!K1}*WKt%oF:B>pTVVY]d~=APU
b0/?c^Wo$7UocBJH|Y@=X[k7a');}elseif($_GET["file"]=="logo.png"){header("Content-Type: image/png");echo
base64_decode('iVBORw0KGgoAAAANSUhEUgAAADkAAAA5BAMAAAB+Np62AAAAMFBMVEUAAACDl60rTnZZdJNziaOerr60vszI0tr8jZH8c3X8SUr309T8Ly78Bgf8r7H6/PpDBKXXAAAAAXRSTlMAQObYZgAAAAlwSFlzAAALEwAACxMBAJqcGAAAAbRJREFUOI3VlM1OwkAQx/sGG0Xh7GwTz7b1AaRwNhqIRy4kPRKjpcc+geEJDHc1chYPfYJ6N7I+gJFQE+UjJIyzS6FqqzeN/A/dtr/Mzsx/PzRtlYSI0fd0Ju5+wDMhHjCTMIqaXoS9QWYw3iLlvRHtLMrwKqDnNLyM4m+lReizCOjXWCgqWdPzvLgJNgnvUGNPV6IVyc7cim2SrHKDMMN+L6DhTKgBDVhqCyPWFW3KwfpqwEOAXUembeYAtn0W3ssErN+RdbxBOcBYowrU2Di8VrEdWcQrx0QjqGlx3m5LUThK4DFRNhGy5lkwp2CVHZ9Qs2ICUY1cGmiUfj7zOnBTyYAdo6a8otjzR0X1UT3uSc97kiqfFzPrMqM39woVZcoUTOhCin7QL1IoJLAOKcrniyCXwUhRboBplTYPSrYJPJ3XLS6Wd8fJqmrqVm2r6vxtvz9T3kigm3bDzPvxxqmn3QDg1l7VcasbtgEpqg+X2133ixlVuTky0Sw7/8eNF+4ncPi1oyFYy4Pk2tz/TPFELrt0w6aX/S93FMPT5OwXUvcbnQl3rWTT1nIy78akqjRbPb0DRTX3Uyvxl2MAAAAASUVORK5CYII=');}exit;}if(preg_match('~^/[-\w.]~',$_SERVER["HTTP_X_FORWARDED_PREFIX"]))$_SERVER["REQUEST_URI"]=$_SERVER["HTTP_X_FORWARDED_PREFIX"].$_SERVER["REQUEST_URI"];define('Adminer\HTTPS',($_SERVER["HTTPS"]&&strcasecmp($_SERVER["HTTPS"],"off"))||ini_bool("session.cookie_secure"));ini_set("session.use_trans_sid",'0');ini_set("arg_separator.output","&");if(!defined("SID")){session_cache_limiter("");session_name("adminer_sid");session_set_cookie_params(0,cookie_path(),"",HTTPS,true);session_start();}if(function_exists("get_magic_quotes_gpc")&&get_magic_quotes_gpc()){$_GET=remove_slashes($_GET,$gd);$_POST=remove_slashes($_POST,$gd);$_COOKIE=remove_slashes($_COOKIE,$gd);}if(function_exists("get_magic_quotes_runtime")&&get_magic_quotes_runtime())set_magic_quotes_runtime(false);if(function_exists('set_time_limit'))set_time_limit(0);ini_set("precision",'16');function
lang($u,$Yf=null){$ta=func_get_args();$ta[0]=$u;return
call_user_func_array('Adminer\lang_format',$ta);}function
lang_format($nj,$Yf=null){if(is_array($nj)){$G=($Yf==1?0:1);$nj=$nj[$G];}$nj=str_replace("'",'’',$nj);$ta=func_get_args();array_shift($ta);$qd=str_replace("%d","%s",$nj);if($qd!=$nj)$ta[0]=format_number($Yf);return
vsprintf($qd,$ta);}define('Adminer\LANG','en');abstract
class
SqlDb{static$instance;static$untrusted=false;var$extension;var$flavor='';var$server_info;var$affected_rows=0;var$info='';var$errno=0;var$error='';protected$multi;abstract
function
attach($N,$V,$Ug);abstract
function
quote($Q);abstract
function
select_db($Kb);abstract
function
query($H,$yj=false);function
multi_query($H){return$this->multi=$this->query($H);}function
store_result(){return$this->multi;}function
next_result(){return
false;}function
inTransaction(){return
false;}}if(extension_loaded('pdo')){abstract
class
PdoDb
extends
SqlDb{protected$pdo;function
dsn($qc,$V,$Ug,array$sg=array()){$sg[\PDO::ATTR_ERRMODE]=\PDO::ERRMODE_SILENT;$sg[\PDO::ATTR_STATEMENT_CLASS]=array('Adminer\PdoResult');try{$this->pdo=new
\PDO($qc,$V,$Ug,$sg);}catch(\Exception$Jc){return$Jc->getMessage();}$this->server_info=@$this->pdo->getAttribute(\PDO::ATTR_SERVER_VERSION);return'';}function
quote($Q){return$this->pdo->quote($Q);}function
query($H,$yj=false){$I=$this->pdo->query($H);$this->error="";if(!$I){list(,$this->errno,$this->error)=$this->pdo->errorInfo();if(!$this->error)$this->error='Unknown error.';return
false;}$this->store_result($I);return$I;}function
store_result($I=null){if(!$I){$I=$this->multi;if(!$I)return
false;}if($I->columnCount()){$I->num_rows=$I->rowCount();return$I;}$this->affected_rows=$I->rowCount();return
true;}function
next_result(){$I=$this->multi;if(!is_object($I))return
false;$I->_offset=0;return@$I->nextRowset();}function
inTransaction(){return$this->pdo->inTransaction();}}class
PdoResult
extends
\PDOStatement{var$_offset=0,$num_rows;function
fetch_assoc(){return$this->fetch_array(\PDO::FETCH_ASSOC);}function
fetch_row(){return$this->fetch_array(\PDO::FETCH_NUM);}private
function
fetch_array($If){$J=$this->fetch($If);return($J?array_map(array($this,'unresource'),$J):$J);}private
function
unresource($X){return(is_resource($X)?stream_get_contents($X):$X);}function
fetch_field(){$K=(object)$this->getColumnMeta($this->_offset++);$U=$K->pdo_type;$K->type=($U==\PDO::PARAM_INT?0:15);$K->charsetnr=($U==\PDO::PARAM_LOB||(isset($K->flags)&&in_array("blob",(array)$K->flags))?63:0);return$K;}function
seek($dg){for($s=0;$s<$dg;$s++)$this->fetch();}}}function
add_driver($t,$D){SqlDriver::$drivers[$t]=$D;}function
get_driver($t){return
SqlDriver::$drivers[$t];}abstract
class
SqlDriver{static$instance;static$drivers=array();static$extensions=array();static$jush;protected$conn;protected$types=array();var$delimiter=";";var$insertFunctions=array();var$editFunctions=array();var$unsigned=array();var$operators=array();var$functions=array();var$grouping=array();var$onActions="RESTRICT|NO ACTION|CASCADE|SET NULL|SET DEFAULT";var$partitionBy=array();var$inout="IN|OUT|INOUT";var$enumLength="'(?:''|[^'\\\\]|\\\\.)*'";var$generated=array();var$primary="";static
function
connect($N,$V,$Ug){$e=new
Db;return($e->attach($N,$V,$Ug)?:$e);}function
__construct(Db$e){$this->conn=$e;}function
types(){return
call_user_func_array('array_merge',array_values($this->types));}function
structuredTypes(){return
array_map('array_keys',$this->types);}function
enumLength(array$l){}function
unconvertFunction(array$l){}function
select($R,array$M,array$Z,array$r,array$E=array(),$z=1,$F=0,$nh=false){$ze=(count($r)<count($M));$H=adminer()->selectQueryBuild($M,$Z,$r,$E,$z,$F);if(!$H)$H="SELECT".limit(($_GET["page"]!="last"&&$z&&$r&&$ze&&JUSH=="sql"?"SQL_CALC_FOUND_ROWS ":"").implode(", ",$M)."\nFROM ".table($R),($Z?"\nWHERE ".implode(" AND ",$Z):"").($r&&$ze?"\nGROUP BY ".implode(", ",$r):"").($E?"\nORDER BY ".implode(", ",$E):""),$z,($F?$z*$F:0),"\n");$zi=microtime(true);$J=$this->conn->query($H,(!$z&&!$nh?1:0));if($nh)echo
adminer()->selectQuery($H,$zi,!$J);return$J;}function
delete($R,$vh,$z=0){$H="FROM ".table($R);return
queries("DELETE".($z?limit1($R,$H,$vh):" $H$vh"));}function
update($R,array$O,$vh,$z=0,$ei="\n"){$Tj=array();foreach($O
as$x=>$X)$Tj[]="$x = $X";$H=table($R)." SET$ei".implode(",$ei",$Tj);return
queries("UPDATE".($z?limit1($R,$H,$vh,$ei):" $H$vh"));}function
insert($R,array$O){return
queries("INSERT INTO ".table($R).($O?" (".implode(", ",array_keys($O)).")\nVALUES (".implode(", ",$O).")":" DEFAULT VALUES").$this->insertReturning($R));}function
insertReturning($R){return"";}function
insertUpdate($R,array$L,array$mh){foreach($L
as$O){$Z=array();foreach($O
as$x=>$X){if(isset($mh[idf_unescape($x)]))$Z[]="$x = $X";}if(!($Z&&$this->update($R,$O," WHERE ".implode(" AND ",$Z))&&$this->conn->affected_rows)&&!$this->insert($R,$O))return
false;}return
true;}function
begin(){return
queries("BEGIN");}function
commit(){return
queries("COMMIT");}function
rollback(){return
queries("ROLLBACK");}function
slowQuery($H,$aj){}function
convertSearch($u,array$X,array$l){return$u;}function
value($X,array$l){return(method_exists($this->conn,'value')?$this->conn->value($X,$l):$X);}function
quoteBinary($Th){return
q($Th);}function
typeName(\stdClass$l){return(isset($l->native_type)?$l->native_type:"");}function
warnings(){}function
tableHelp($D,$Ce=false){}function
inheritsFrom($R){return
array();}function
inheritedTables($R){return
array();}function
partitionsInfo($R){return
array();}function
hasCStyleEscapes(){return
false;}function
lineComment(){return"--";}function
engines(){return
array();}function
supportsIndex(array$S){return!is_view($S);}function
supportsAlterIndex(array$S){return
true;}function
indexAlgorithms(array$Ki){return
array();}function
indexOpclasses(){return
array();}function
checkConstraints($R){return
get_key_vals("SELECT c.CONSTRAINT_NAME, CHECK_CLAUSE
FROM INFORMATION_SCHEMA.CHECK_CONSTRAINTS c
JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS t
	ON c.CONSTRAINT_SCHEMA = t.CONSTRAINT_SCHEMA AND c.CONSTRAINT_NAME = t.CONSTRAINT_NAME".($this->conn->flavor=='maria'?" AND c.TABLE_NAME = ".q($R):"")."
WHERE c.CONSTRAINT_SCHEMA = ".q($_GET["ns"]!=""?$_GET["ns"]:DB)."
AND t.TABLE_NAME = ".q($R).(JUSH=="pgsql"?"
AND CHECK_CLAUSE NOT LIKE '% IS NOT NULL'":""),$this->conn);}function
allFields(){$J=array();if(DB!=""){foreach(get_rows("SELECT c.TABLE_NAME AS tab, c.COLUMN_NAME AS field, c.IS_NULLABLE AS nullable,
	c.DATA_TYPE AS type, c.CHARACTER_MAXIMUM_LENGTH AS length,
	".(JUSH=='sql'?"c.COLUMN_KEY = 'PRI'":"k.COLUMN_NAME")." AS ".idf_escape("primary")."
FROM INFORMATION_SCHEMA.COLUMNS c".(JUSH=='sql'?"":"
LEFT JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS t ON c.TABLE_SCHEMA = t.TABLE_SCHEMA AND c.TABLE_NAME = t.TABLE_NAME AND t.CONSTRAINT_TYPE = 'PRIMARY KEY'
LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE k
	ON t.CONSTRAINT_SCHEMA = k.CONSTRAINT_SCHEMA AND t.CONSTRAINT_NAME = k.CONSTRAINT_NAME AND c.TABLE_SCHEMA = k.TABLE_SCHEMA AND c.TABLE_NAME = k.TABLE_NAME AND c.COLUMN_NAME = k.COLUMN_NAME")."
WHERE c.TABLE_SCHEMA = ".q($_GET["ns"]!=""?$_GET["ns"]:DB)."
ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION",$this->conn)as$K){$K["null"]=($K["nullable"]=="YES");$J[$K["tab"]][]=$K;}}return$J;}}class
Adminer{static$instance;var$error='';function
name(){return"<a href='https://www.adminer.org/'".target_blank()." id='h1'><img src='".h(preg_replace("~\\?.*~","",ME)."?file=logo.png&version=6.0.0")."' width='24' height='24' alt='' id='logo'>Adminer</a>";}function
credentials(){return
array(SERVER,$_GET["username"],get_password());}function
connectSsl(){}function
permanentLogin($g=false){return
password_file($g);}function
bruteForceKey(){return$_SERVER["REMOTE_ADDR"];}function
serverName($N){return
h($N);}function
database(){return
DB;}function
databases($ld=true){return
get_databases($ld);}function
pluginsLinks(){}function
operators(){return
driver()->operators;}function
schemas(){$J=schemas();if($_GET["ns"]!=""&&!in_array($_GET["ns"],$J))array_unshift($J,$_GET["ns"]);return$J;}function
queryTimeout(){return
2;}function
afterConnect(){}function
headers(){}function
csp(array$Bb){return$Bb;}function
verifyVersion(){return
true;}function
head($Gb=null){return
true;}function
bodyClass(){echo" adminer";}function
css(){$J=array();foreach(array("","-dark")as$If){$n="adminer$If.css";if(file_exists($n)){$cd=file_get_contents($n);$J["$n?v=".crc32($cd)]=($If?"dark":(preg_match('~prefers-color-scheme:\s*dark~',$cd)?'':'light'));}}return$J;}function
loginForm(){echo"<table class='layout'>\n",adminer()->loginFormField('driver','<tr><th>'.'System'.'<td>',input_hidden("auth[driver]","server")."MySQL / MariaDB"),adminer()->loginFormField('server','<tr><th>'.'Server'.'<td>',"<input name='auth[server]' value='".h(SERVER)."' title='".'hostname[:port] or :socket'."' placeholder='localhost' autocapitalize='off'>"),adminer()->loginFormField('username','<tr><th>'.'Username'.'<td>','<input name="auth[username]" id="username" autofocus value="'.h($_GET["username"]).'" autocomplete="username" autocapitalize="off">'),adminer()->loginFormField('password','<tr><th>'.'Password'.'<td>','<input type="password" name="auth[password]" autocomplete="current-password">'),adminer()->loginFormField('db','<tr><th>'.'Database'.'<td>','<input name="auth[db]" value="'.h($_GET["db"]).'" autocapitalize="off">'),"</table>\n","<p><input type='submit' value='".'Login'."'>\n",checkbox("auth[permanent]",1,$_COOKIE["adminer_permanent"],'Permanent login')."\n";}function
loginFormField($D,$Od,$Y){return$Od.$Y."\n";}function
login($ef,$Ug){if($Ug==""||!password_required())return
sprintf('Adminer does not support accessing a database without a password, <a href="https://www.adminer.org/en/password/"%s>more information</a>.',target_blank());return
true;}function
tableName(array$Ki){return
h($Ki["Name"]);}function
fieldName(array$l,$E=0){$U=$l["full_type"].($l["null"]?" NULL":"");$jb=$l["comment"];return'<span title="'.h($U.($jb!=""?($U?": ":"").$jb:'')).'">'.h($l["field"]).'</span>';}function
commentValue($U,$jb){if($jb==""||$U=='TABLE'||$U=='COLUMN')return
h($jb);$jh=function($Th){return
preg_replace('~^~m','<tr>',preg_replace('~\|~','<td>',preg_replace('~\|$~m',"",rtrim($Th))));};$R='(\+--[-+]+\+\n)';$K='(\| .* \|\n)';return"<pre>\n".preg_replace_callback("~^$R?$K$R?($K*)$R?~m",function($B)use($jh){$jd=$jh($B[2]);return"<table>\n".($B[1]?"<thead>$jd<tbody>\n":$jd).$jh($B[4])."\n</table>";},preg_replace('~(\n(    -|mysql)&gt; )(.+)~',"\\1<code class='jush-sql'>\\3</code>",preg_replace('~(.+)\n---+\n~',"<b>\\1</b>\n",h($jb))))."</pre>\n";}function
commentInput($U,$b,$jb){$Y=h($jb);return(preg_match('~\n~',$Y)?"<textarea$b rows='2' cols='".($U=='TABLE'?20:30)."' style='vertical-align: bottom;'>\n$Y</textarea>":"<input$b value='$Y'>");}function
selectLinks(array$Ki,$O=""){$D=$Ki["Name"];echo'<p class="links">';$af=array("select"=>'Select data');if(support("table")||support("indexes"))$af["table"]='Show structure';$Ce=false;if(support("table")){$Ce=is_view($Ki);if($Ce){if(support("view"))$af["view"]='Alter view';}elseif(function_exists('Adminer\alter_table')&&$D!="")$af["create"]='Alter table';}if($O!==null)$af["edit"]='New item';foreach($af
as$x=>$X)echo" <a href='".h(ME)."$x=".url_escape($D).($x=="edit"?$O:"")."'".bold(isset($_GET[$x])).">$X</a>";echo
doc_link(array(JUSH=>driver()->tableHelp($D,$Ce)),"?"),"\n";}function
foreignKeys($R){return
foreign_keys($R);}function
backwardKeys($R,$Ji){return
array();}function
backwardKeysPrint(array$Da,array$K){}function
selectQuery($H,$zi,$Vc=false){$J="\n";if(!$Vc&&($bk=driver()->warnings())){$t="warnings";$J=", <a href='#$t' class='toggle'>".'Warnings'."</a>"."$J<div id='$t' class='hidden'>\n$bk</div>\n";}return"<p><code class='jush-".JUSH."'>".h(str_replace("\n"," ",$H))."</code> <span class='time'>(".format_time($zi).")</span>".(support("sql")?" <a href='".h(ME)."sql=".url_escape($H)."' class='hover'>".'Edit'."</a>":"").$J;}function
sqlCommandQuery($H){return
shorten_utf8(trim($H),1000);}function
sqlPrintAfter(){}function
rowDescription($R){return"";}function
rowDescriptions(array$L,array$od){return$L;}function
selectLink($X,array$l){}function
selectVal($X,$_,array$l,$Cg){$J=($X===null?"<i>NULL</i>":(preg_match("~char|binary|boolean~",$l["type"])&&!preg_match("~var~",$l["type"])?"<code>$X</code>":(preg_match('~^jsonb?$~',$l["full_type"])?"<code class='jush-json'>$X</code>":$X)));if(is_blob($l)&&!is_utf8($X))$J="<i>".lang_format(array('%d byte','%d bytes'),strlen($Cg))."</i>";return($_?"<a href='".h($_)."'".(is_url($_)?target_blank():"").">$J</a>":$J);}function
editVal($X,array$l){return$X;}function
config(){return
array();}function
tableStructurePrint(array$m,$Ki=null){echo"<div class='scrollable'>\n","<table class='nowrap odds'>\n","<thead><tr><th>".'Column'."<td>".'Type'.(support("comment")?"<td>".'Comment':"")."<tbody>\n";$Bi=driver()->structuredTypes();foreach($m
as$l){echo"<tr><th>".h($l["field"]);$U=h($l["full_type"]);$fb=h($l["collation"]);echo"<td><span title='$fb'>".(in_array($U,(array)$Bi['User types'])?"<a href='".h(ME.'type='.url_escape($U))."'>$U</a>":$U.($fb&&isset($Ki["Collation"])&&$fb!=$Ki["Collation"]?" $fb":""))."</span>",($l["null"]?" <i>NULL</i>":""),($l["auto_increment"]?" <i>".'Auto Increment'."</i>":""),(isset($l["default"])?" <span title='".'Default value'."'>[<b>".($l["generated"]?"<code class='jush-".JUSH."'>".shorten_utf8(preg_replace('~\s+~',' ',ltrim($l["default"])),80,"</code>"):h($l["default"]))."</b>]</span>":""),(support("comment")?"<td>".adminer()->commentValue('COLUMN',$l["comment"]):""),"\n";}echo"</table>\n","</div>\n";}function
tableIndexesPrint(array$w,array$Ki){$Lg=false;foreach($w
as$D=>$v)$Lg|=!!$v["partial"];echo"<table>\n";$Pb=first(driver()->indexAlgorithms($Ki));foreach($w
as$D=>$v){ksort($v["columns"]);$nh=array();foreach($v["columns"]as$x=>$X)$nh[]="<i>".h($X)."</i>".($v["lengths"][$x]?"(".h($v["lengths"][$x]).")":"").($v["descs"][$x]?" DESC":"");echo"<tr title='".h($D)."'>","<th>".h($v["type"]).($Pb&&$v['algorithm']!=$Pb?" (".h($v['algorithm']).")":""),"<td>".implode(", ",$nh);if($Lg)echo"<td>".($v['partial']?"<code class='jush-".JUSH."'>WHERE ".h($v['partial']):"");echo"\n";}echo"</table>\n";}function
selectColumnsPrint(array$M,array$d){print_fieldset("select",'Select',$M);$s=0;$M[""]=array();foreach($M
as$x=>$X){$X=idx($_GET["columns"],$x,array());$c=select_input(" name='columns[$s][col]' data-default=''".on('change',($x!==""?'selectFieldChange':'selectAddRow')),$d,$X["col"]);echo"<div>".(driver()->functions||driver()->grouping?html_select("columns[$s][fun]",array(-1=>"")+array_filter(array('Functions'=>driver()->functions,'Aggregation'=>driver()->grouping)),$X["fun"]," data-default=''".on('change',($x!==""?'helpClose':'selectFunAddRow')).on_help_value(' (.*)|$','($1)'))."($c)":$c)."</div>\n";$s++;}echo"</div></fieldset>\n";}function
selectSearchPrint(array$Z,array$d,array$w){print_fieldset("search",'Search',$Z);foreach($w
as$s=>$v){if($v["type"]=="FULLTEXT")echo"<div>(<i>".implode("</i>, <i>",array_map('Adminer\h',$v["columns"]))."</i>) AGAINST"," <input type='search' name='fulltext[$s]' value='".h(idx($_GET["fulltext"],$s))."' data-default=''".on('input','selectFieldChange').">",(JUSH=='sql'?checkbox("boolean[$s]",1,isset($_GET["boolean"][$s]),"BOOL"):''),"</div>\n";}$pg=adminer()->operators();foreach(array_merge((array)$_GET["where"],array(array()))as$s=>$X){if(!$X||("$X[col]$X[val]"!=""&&in_array($X["op"],$pg)))echo"<div>".select_input(" name='where[$s][col]' data-default=''".on('change',($X?'selectFieldChange':'selectAddRow')),$d,$X["col"],"(".'anywhere'.")"),html_select("where[$s][op]",$pg,$X["op"]," data-default='".h(first($pg))."'".on('change','selectFirstChange')),"<input type='search' name='where[$s][val]' value='".h($X["val"])."' data-default=''".on('input','selectFirstChange').on('keydown','selectSearchKeydown').on('search','selectSearchSearch').">","</div>\n";}echo"</div></fieldset>\n";}function
selectOrderPrint(array$E,array$d,array$w){print_fieldset("sort",'Sort',$E);$s=0;foreach((array)$_GET["order"]as$x=>$X){if($X!=""){echo"<div>".select_input(" name='order[$s]' data-default=''".on('change','selectFieldChange'),$d,$X),checkbox("desc[$s]",1,isset($_GET["desc"][$x]),'descending')."</div>\n";$s++;}}echo"<div>".select_input(" name='order[$s]' data-default=''".on('change','selectAddRow'),$d),checkbox("desc[$s]",1,false,'descending')."</div>\n","</div></fieldset>\n";}function
selectLimitPrint($z){echo"<fieldset><legend>".'Limit'."</legend><div>","<input type='number' name='limit' class='size' value='".h($z?:"")."' data-default='50'".on('input','selectFieldChange').">","</div></fieldset>\n";}function
selectLengthPrint($Yi){echo"<fieldset><legend>".'Text length'."</legend><div>","<input type='number' name='text_length' class='size' value='".h($Yi)."' data-default='100'>","</div></fieldset>\n";}function
selectActionPrint(array$w){echo"<fieldset><legend>".'Action'."</legend><div>","<input type='submit' value='".'Select'."'>"," <span id='noindex' title='".'Full table scan'."'></span>","<script".nonce().">\n","const indexColumns = ";$d=array();foreach($w
as$v){$Fb=reset($v["columns"]);if($v["type"]!="FULLTEXT"&&$Fb)$d[$Fb]=1;}$d[""]=1;foreach($d
as$x=>$X)json_row($x);echo";\n","selectFieldChange.call(qs('#form')['select']);\n","</script>\n","</div></fieldset>\n";}function
selectCommandPrint(){return!information_schema(DB);}function
selectImportPrint(){return!information_schema(DB);}function
selectEmailPrint(array$wc,array$d){}function
selectColumnsProcess(array$d,array$w){$M=array();$r=array();foreach((array)$_GET["columns"]as$x=>$X){if($X["fun"]=="count"||($X["col"]!=""&&(!$X["fun"]||in_array($X["fun"],driver()->functions)||in_array($X["fun"],driver()->grouping)))){$M[$x]=apply_sql_function($X["fun"],($X["col"]!=""?idf_escape($X["col"]):"*"));if(!in_array($X["fun"],driver()->grouping))$r[]=$M[$x];}}return
array($M,$r);}function
selectSearchProcess(array$m,array$w){$J=array();foreach($w
as$s=>$v){if($v["type"]=="FULLTEXT"&&idx($_GET["fulltext"],$s)!="")$J[]="MATCH (".implode(", ",array_map('Adminer\idf_escape',$v["columns"])).") AGAINST (".q($_GET["fulltext"][$s]).(isset($_GET["boolean"][$s])?" IN BOOLEAN MODE":"").")";}$pg=adminer()->operators();foreach((array)$_GET["where"]as$x=>$X){$X+=array("col"=>"","op"=>first($pg),"val"=>"");$_GET["where"][$x]=$X;$db=$X["col"];if("$db$X[val]"!=""&&in_array($X["op"],$pg)){if($X["op"]=="SQL"&&(!$_POST||!verify_token()))SqlDb::$untrusted=true;$ob=array();foreach(($db!=""?array($db=>$m[$db]):$m)as$D=>$l){$kh="";$nb=" $X[op]";if(preg_match('~IN$~',$X["op"])){$de=process_length($X["val"]);$nb
.=" ".($de!=""?$de:"(NULL)");}elseif($X["op"]=="SQL")$nb=" $X[val]";elseif(preg_match('~^(I?LIKE) %%$~',$X["op"],$B))$nb=" $B[1] ".adminer()->processInput($l,"%$X[val]%");elseif($X["op"]=="FIND_IN_SET"){$kh="$X[op](".q($X["val"]).", ";$nb=")";}elseif(!preg_match('~NULL$~',$X["op"]))$nb
.=" ".adminer()->processInput($l,$X["val"]);if($db!=""||(isset($l["privileges"]["where"])&&(preg_match('~^[-\d.'.(preg_match('~IN$~',$X["op"])?',':'').']+$~',$X["val"])||!preg_match('~'.number_type().'|bit~',$l["type"]))&&(!preg_match("~[\x80-\xFF]~",$X["val"])||preg_match('~char|text|enum|set~',$l["type"]))&&(!preg_match('~date|timestamp~',$l["type"])||preg_match('~^\d+-\d+-\d+~',$X["val"]))))$ob[]=$kh.driver()->convertSearch(idf_escape($D),$X,$l).$nb;}$J[]=(count($ob)==1?$ob[0]:($ob?"(".implode(" OR ",$ob).")":"1 = 0"));}}return$J;}function
selectOrderProcess(array$m,array$w){$J=array();foreach((array)$_GET["order"]as$x=>$X){if($X!="")$J[]=(preg_match('~^((COUNT\(DISTINCT |[A-Z0-9_]+\()(`(?:[^`]|``)+`|"(?:[^"]|"")+")\)|COUNT\(\*\))$~',$X)?$X:idf_escape($X)).(isset($_GET["desc"][$x])?" DESC".(JUSH=='pgsql'&&idx($m[$X],"null")?" NULLS LAST":""):"");}return$J;}function
selectLimitProcess(){return(isset($_GET["limit"])?intval($_GET["limit"]):50);}function
selectLengthProcess(){return(isset($_GET["text_length"])?"$_GET[text_length]":"100");}function
selectEmailProcess(array$Z,array$od){return
false;}function
selectQueryBuild(array$M,array$Z,array$r,array$E,$z,$F){return"";}function
messageQuery($H,$Zi,$Vc=false){restart_session();$Rd=&get_session("queries");if(!idx($Rd,$_GET["db"]))$Rd[$_GET["db"]]=array();if(strlen($H)>1e6)$H=preg_replace('~[\x80-\xFF]+$~','',substr($H,0,1e6))."\n…";$Rd[$_GET["db"]][]=array($H,time(),$Zi);$vi="sql-".count($Rd[$_GET["db"]]);$J="<a href='#$vi' class='toggle'>".'SQL command'."</a> ".copy_icon()."\n";if(!$Vc&&($bk=driver()->warnings())){$t="warnings-".count($Rd[$_GET["db"]]);$J="<a href='#$t' class='toggle'>".'Warnings'."</a>, $J<div id='$t' class='hidden'>\n$bk</div>\n";}return" <span class='time'>".@date("H:i:s")."</span>"." $J<div id='$vi' class='hidden'><pre><code class='jush-".JUSH."'>".shorten_utf8($H,1e4)."</code></pre>".($Zi?" <span class='time'>($Zi)</span>":'').(support("sql")?'<p><a href="'.h(str_replace("db=".url_escape(DB),"db=".url_escape($_GET["db"]),ME).'sql=&history='.(count($Rd[$_GET["db"]])-1)).'">'.'Edit'.'</a>':'').'</div>';}function
editRowPrint($R,array$m,$K,$Fj){}function
editFunctions(array$l){$J=($l["null"]?"NULL/":"");$Ld=isset($_GET["select"])||where($_GET);foreach(array(driver()->insertFunctions,driver()->editFunctions)as$x=>$xd){if(!$x||(!isset($_GET["call"])&&$Ld)){foreach($xd
as$Wg=>$X){if(!$Wg||preg_match("~$Wg~",$l["type"]))$J
.="/$X";}}if($x&&$xd&&!preg_match('~set|bool~',$l["type"])&&!is_blob($l))$J
.="/SQL";}if($l["auto_increment"]&&!$Ld)$J='Auto Increment';return
explode("/",$J);}function
editInput($R,array$l,$b,$Y){if($l["type"]=="enum")return(isset($_GET["select"])?"<label><input type='radio'$b value='orig' checked><i>".'original'."</i></label> ":"").enum_input("radio",$b,$l,$Y,"NULL");return"";}function
editHint($R,array$l,$Y){return"";}function
processInput(array$l,$Y,$q=""){if($q=="SQL")return$Y;$D=$l["field"];$J=q($Y);if(preg_match('~^(now|getdate|uuid)$~',$q))$J="$q()";elseif(preg_match('~^current_(date|timestamp)$~',$q))$J=$q;elseif(preg_match('~^([+-]|\|\|)$~',$q))$J=idf_escape($D)." $q $J";elseif(preg_match('~^[+-] interval$~',$q))$J=idf_escape($D)." $q ".(preg_match("~^(\\d+|'[0-9.: -]') [A-Z_]+\$~i",$Y)&&JUSH!="pgsql"?$Y:$J);elseif(preg_match('~^(addtime|subtime|concat)$~',$q))$J="$q(".idf_escape($D).", $J)";elseif(preg_match('~^(md5|sha1|password|encrypt)$~',$q))$J="$q($J)";return
unconvert_field($l,$J);}function
dumpOutput(){$J=array('text'=>'open','file'=>'save');if(function_exists('gzencode'))$J['gz']='gzip';return$J;}function
dumpFormat(){return(support("dump")?array('sql'=>'SQL'):array())+array('csv'=>'CSV,','csv;'=>'CSV;','tsv'=>'TSV');}function
dumpDatabase($i){}function
dumpTable($R,$Ci,$Ce=0){if($_POST["format"]!="sql"){echo"\xef\xbb\xbf";if($Ci)dump_csv(array_keys(fields($R)));}else{if($Ce==2){$m=array();foreach(fields($R)as$D=>$l)$m[]=idf_escape($D)." $l[full_type]";$g="CREATE TABLE ".table($R)." (".implode(", ",$m).")";}else$g=create_sql($R,$_POST["auto_increment"],$Ci);set_utf8mb4($g);if($Ci&&$g){if(($Ci=="DROP+CREATE"&&!function_exists('Adminer\drop_sql'))||$Ce==1)echo"DROP ".($Ce==2?"VIEW":"TABLE")." IF EXISTS ".table($R).";\n";if($Ce==1)$g=remove_definer($g);echo"$g;\n\n";}}}function
dumpData($R,$Ci,$H,array$M=array(),array$Z=array(),array$r=array(),array$E=array()){if($Ci){$of=(JUSH=="sqlite"?0:1048576);$m=array();$Zd=false;if($_POST["format"]=="sql"){if($Ci=="TRUNCATE+INSERT"&&!function_exists('Adminer\truncate_all_sql'))echo
truncate_sql($R).";\n";$m=fields($R);if(JUSH=="mssql"){foreach($m
as$l){if($l["auto_increment"]){echo"SET IDENTITY_INSERT ".table($R)." ON;\n";$Zd=true;break;}}}}$I=($H!=""?connection()->query($H,1):driver()->select($R,($M?:array("*")),$Z,$r,$E,0));if($I){$re="";$Na="";$Ie=array();$yd=array();$Ei="";$Yc=($R!=''?'fetch_assoc':'fetch_row');$yb=0;while($K=$I->$Yc()){if(!$Ie){$Tj=array();foreach($K
as$X){$l=$I->fetch_field();if(idx($m[$l->name],'generated')){$yd[$l->name]=true;continue;}$Ie[]=$l->name;$x=idf_escape($l->name);$Tj[]="$x = VALUES($x)";}$Ei=($Ci=="INSERT+UPDATE"?"\nON DUPLICATE KEY UPDATE ".implode(", ",$Tj):"").";\n";}if($_POST["format"]!="sql"){if($Ci=="table"){dump_csv($Ie);$Ci="INSERT";}dump_csv($K);}else{if(!$re)$re="INSERT INTO ".table($R)." (".implode(", ",array_map('Adminer\idf_escape',$Ie)).") VALUES";foreach($K
as$x=>$X){if($yd[$x]){unset($K[$x]);continue;}$l=$m[$x];$K[$x]=($X===null?"NULL":($X===false?0:unconvert_field($l,preg_match(number_type(),$l["type"])&&!preg_match('~\[~',$l["full_type"])&&is_numeric($X)?$X:(!is_blob($l)||is_utf8($X)?q($X):driver()->quoteBinary($X)))));}$Th=($of?"\n":" ")."(".implode(",\t",$K).")";if(!$Na)$Na=$re.$Th;elseif(JUSH=='mssql'?$yb%1000!=0:strlen($Na)+4+strlen($Th)+strlen($Ei)<$of)$Na
.=",$Th";else{echo$Na.$Ei;$Na=$re.$Th;}}$yb++;}if($Na)echo$Na.$Ei;}elseif($_POST["format"]=="sql")echo"-- ".str_replace("\n"," ",connection()->error)."\n";if($Zd)echo"SET IDENTITY_INSERT ".table($R)." OFF;\n";}}function
dumpFilename($Yd){return
friendly_url($Yd!=""?$Yd:(SERVER?:"localhost"));}function
dumpHeaders($Yd,$Kf=false){$Fg=$_POST["output"];$Qc=(preg_match('~sql~',$_POST["format"])?"sql":($Kf?"tar":"csv"));header("Content-Type: ".($Fg=="gz"?"application/x-gzip":($Qc=="tar"?"application/x-tar":($Qc=="sql"||$Fg!="file"?"text/plain":"text/csv")."; charset=utf-8")));if($Fg=="gz"){ob_start(function($Q){return
gzencode($Q);},1e6);}return$Qc;}function
dumpFooter(){if($_POST["format"]=="sql")echo"-- ".gmdate("Y-m-d H:i:s e")."\n";}function
importServerPath(){return"adminer.sql";}function
importPrint(){}function
importProcess(){return
false;}function
homepage(){echo'<p class="links">'.($_GET["ns"]==""&&support("database")?'<a href="'.h(ME).'database=">'.'Alter database'."</a>\n":""),(support("scheme")?"<a href='".h(ME)."scheme='>".($_GET["ns"]!=""?'Alter schema':'Create schema')."</a>\n":""),($_GET["ns"]!==""?'<a href="'.h(ME).'schema=">'.'Database schema'."</a>\n":""),(support("privileges")?"<a href='".h(ME)."privileges='>".'Privileges'."</a>\n":"");if($_GET["ns"]!=="")echo(support("routine")?"<a href='#routines'>".'Routines'."</a>\n":""),(support("sequence")?"<a href='#sequences'>".'Sequences'."</a>\n":""),(support("type")?"<a href='#user-types'>".'User types'."</a>\n":""),(support("event")?"<a href='#events'>".'Events'."</a>\n":"");return
true;}function
navigation($Hf){echo"<h1>".adminer()->name()." <span class='version'>".VERSION;$Tf=$_COOKIE["adminer_version"];echo" <a href='https://www.adminer.org/#download'".target_blank()." id='version'>".(version_compare(VERSION,$Tf)<0?h($Tf):"").version_iframe()."</a>","</span></h1>\n";if($Hf=="auth"){$Fg="";foreach((array)$_SESSION["pwds"]as$Vj=>$gi){foreach($gi
as$N=>$Pj){$D=h(get_setting("vendor-$Vj-$N")?:get_driver($Vj));foreach($Pj
as$V=>$Ug){if($D&&$Ug!==null){$Nb=$_SESSION["db"][$Vj][$N][$V];foreach(($Nb?array_keys($Nb):array(""))as$i)$Fg
.="<li><a href='".h(auth_url($Vj,$N,$V,$i))."'>($D) ".h("$V@").($N!=""?adminer()->serverName($N):"").h($i!=""?" - $i":"")."</a>\n";}}}}if($Fg)echo"<ul id='logins'".on('mouseover','menuOver').on('mouseout','menuOut').">\n$Fg</ul>\n";}else{$T=array();if($_GET["ns"]!==""&&!$Hf&&DB!=""){connection()->select_db(DB);$T=table_status('',true);}adminer()->syntaxHighlighting($T);adminer()->databasesPrint($Hf);$ga=array();if(DB==""||!$Hf){if(support("sql")){$ga['sql']="<a href='".h(ME)."sql='".bold(isset($_GET["sql"])&&!isset($_GET["import"])).">".'SQL command'."</a>";$ga['import']="<a href='".h(ME)."import='".bold(isset($_GET["import"])).">".'Import'."</a>";}$ga['dump']="<a href='".h(ME)."dump=".url_escape(isset($_GET["table"])?$_GET["table"]:$_GET["select"])."' id='dump'".bold(isset($_GET["dump"])).">".'Export'."</a>";}$ee=$_GET["ns"]!==""&&!$Hf&&DB!="";if($ee&&function_exists('Adminer\alter_table'))$ga['create']='<a href="'.h(ME).'create="'.bold($_GET["create"]==="").">".'Create table'."</a>";$ga=adminer()->menuActions($ga,$Hf);echo($ga?"<p class='links'>\n".implode("\n",$ga)."\n":"");if($ee){if($T)adminer()->tablesPrint($T);else
echo"<p class='message'>".'No tables.'."</p>\n";}}}function
syntaxHighlighting(array$T){echo
script_src(preg_replace("~\\?.*~","",ME)."?file=jush.js&version=6.0.0",true);if(support("sql")){$Ge="adminer-plugins/jush-".JUSH.".js";echo(file_exists($Ge)?script_src($Ge,true):""),"<script".nonce().">\n";if($T){$af=array();foreach($T
as$R=>$U)$af[]=js_escape_re($R);echo"var jushLinks = { ".JUSH.":";json_row(js_escape(ME).(support("table")?"table":"select").'=$&','/\b(?<!\$)('.implode('|',$af).')(?!\$)\b/g',false);$xi=array("sql","check","event","procedure","trigger","view","type","table","processlist");if(support("routine")&&array_intersect_key($_GET,array_flip($xi))){foreach(routines()as$K)json_row(js_escape(ME).'function='.url_escape($K["SPECIFIC_NAME"]).'&name=$&','/\b'.js_escape_re($K["ROUTINE_NAME"]).'(?=["`\]]?\()/g',false);}json_row('');echo"};\n";foreach(array("bac","bra","sqlite_quo","mssql_bra")as$X)echo"jushLinks.$X = jushLinks.".JUSH.";\n";if(isset($_GET["sql"])||isset($_GET["trigger"])||isset($_GET["check"])){$Pi=array_fill_keys(array_keys($T),array());foreach(driver()->allFields()as$R=>$m){foreach($m
as$l)$Pi[$R][]=$l["field"];}echo"addEventListener('DOMContentLoaded', () => { autocompleter = jush.autocompleteSql('".idf_escape("")."', ".json_encode($Pi)."); });\n";}}echo"</script>\n";}echo
script("syntaxHighlighting('".(preg_match('~^\d\.?\d~',connection()->server_info,$B)?$B[0]:"")."', '".connection()->flavor."');");}function
databasesPrint($Hf){$h=adminer()->databases();if(DB&&$h&&!in_array(DB,$h))array_unshift($h,DB);echo"<form action=''>\n<p id='dbs'>\n";hidden_fields_get();$Lb=on('mousedown','dbMouseDown').on('change','dbChange');echo"<label title='".'Database'."'>".'DB'.": ".($h?html_select("db",array(""=>"")+$h,DB,$Lb):"<input name='db' value='".h(DB)."' autocapitalize='off' size='19'>\n")."</label>","<input type='submit' value='".'Use'."'".($h?" class='hidden'":"").">\n";foreach(array("import","sql","schema","dump","privileges")as$X){if(isset($_GET[$X])){echo
input_hidden($X);break;}}echo"</p></form>\n";}function
menuActions(array$ga,$Hf){return$ga;}function
tablesPrint(array$T){echo"<ul id='tables'".on('mouseover','menuOver').on('mouseout','menuOut').">";foreach($T
as$R=>$P){$R="$R";$D=adminer()->tableName($P);if($D!=""&&!$P["partition"])echo'<li><a href="'.h(ME).'select='.url_escape($R).'"'.bold($_GET["select"]==$R||$_GET["edit"]==$R,"select hover")." title='".'Select data'."'>".'select'."</a> ",(support("table")||support("indexes")?'<a href="'.h(ME).'table='.url_escape($R).'"'.bold(in_array($R,array($_GET["table"],$_GET["create"],$_GET["indexes"],$_GET["foreign"],$_GET["trigger"],$_GET["check"],$_GET["view"])),(is_view($P)?"view":"structure"))." title='".'Show structure'."'>$D</a>":"<span>$D</span>")."\n";}echo"</ul>\n";}function
showVariables(){return
show_variables();}function
showStatus(){return
show_status();}function
processList(){return
process_list();}function
killProcess($t){return
kill_process($t);}}class
Plugins{private
static$append=array('dumpFormat'=>true,'dumpOutput'=>true,'editRowPrint'=>true,'editFunctions'=>true,'config'=>true);var$plugins;var$drivers=array();var$driverFiles=array();var$error='';private$hooks=array();function
__construct($ch){$mc=SqlDriver::$drivers;$Pd=" href='https://www.adminer.org/plugins/#use'".target_blank();if($ch===null){$ch=array();$Ha="adminer-plugins";if(is_dir($Ha)){foreach(glob("$Ha/*.php")as$n){$dd=SqlDriver::$drivers;$this->includeOnce($n);foreach(array_diff_key(SqlDriver::$drivers,$dd)as$t=>$D)$this->driverFiles[$t]=$n;}}if(file_exists("$Ha.php")){$ge=$this->includeOnce("$Ha.php");if(is_array($ge)){foreach($ge
as$x=>$ah)$ch[is_object($ah)?get_class($ah):$x]=$ah;}else$this->error
.=sprintf('%s must <a%s>return an array</a>.',"<b>$Ha.php</b>",$Pd)."<br>";}foreach(get_declared_classes()as$bb){if(!$ch[$bb]&&(preg_match('~^Adminer\w~i',$bb)||is_subclass_of($bb,'Adminer\Plugin'))){$Dh=new
\ReflectionClass($bb);$qb=$Dh->getConstructor();if($qb&&$qb->getNumberOfRequiredParameters())$this->error
.=sprintf('<a%s>Configure</a> %s in %s.',$Pd,"<b>$bb</b>","<b>$Ha.php</b>")."<br>";else$ch[$bb]=new$bb;}}}$ve=array_filter($ch,function($ah){return!is_object($ah);});if($ve){$this->error
.=sprintf('Every plugin must <a%s>be an object</a>.',$Pd)."<br>";$ch=array_diff_key($ch,$ve);}$this->drivers=array_diff_key(SqlDriver::$drivers,$mc);$this->plugins=$ch;$ha=new
Adminer;$ch[]=$ha;$Dh=new
\ReflectionObject($ha);foreach($Dh->getMethods()as$Ff){foreach($ch
as$ah){$D=$Ff->getName();if(method_exists($ah,$D))$this->hooks[$D][]=$ah;}}}function
includeOnce($n){return
include_once"./$n";}static
function
checksum($n){$cd=str_replace("\r","",file_get_contents($n));$cd=preg_replace('~\n\tprotected \$translations = array\(.*?\n\t\);~s','',$cd);return
dechex(crc32($cd));}function
checksums(){$ed=array_values($this->driverFiles);foreach($this->plugins
as$ah){$Dh=new
\ReflectionObject($ah);$ed[]=$Dh->getFileName();}$J=array();foreach($ed
as$n)$J[basename($n,'.php')]=self::checksum($n);return$J;}static
function
officialChecksums(){return
array('adminer.js'=>'a0599090','backward-keys'=>'afce3b7d','before-unload'=>'48618ca0','config'=>'f49cc617','dark-switcher'=>'3d490dea','database-hide'=>'90c6c0dc','designs'=>'56f1c186','dump-alter'=>'d078b2db','dump-bz2'=>'f0d0e336','dump-date'=>'adc7f1c7','dump-json'=>'767dd321','dump-xml'=>'9f039895','dump-zip'=>'93817d96','edit-foreign'=>'8c874a58','edit-textarea'=>'a24c3cc','editor-setup'=>'a7dc3a37','editor-views'=>'5c12b185','enum-option'=>'a2563959','file-upload'=>'235eaa7a','foreign-system'=>'ebb4c654','frames'=>'b0e1d11a','highlight-codemirror'=>'f1a34275','highlight-monaco'=>'6a92cc58','highlight-prism'=>'4c12cf3','import-csv'=>'1d174088','login-ip'=>'b4766b62','login-otp'=>'62c517c0','login-passkey'=>'f69f2f06','login-password-less'=>'97c37010','login-reverse-proxy'=>'7bb63f11','login-servers'=>'f9ac2f28','login-ssl'=>'6ed147bc','login-table'=>'7b15c3cd','menu-links'=>'f1f86a60','remote-color'=>'33a766c2','row-numbers'=>'eec8698c','select-email'=>'ead22272','select-image'=>'f55c0231','slugify'=>'4d5adde6','sql-gemini'=>'fabc3537','sql-log'=>'b4355039','table-indexes-structure'=>'a90cc0c9','table-structure'=>'a8458e02','tables-filter'=>'f8f51976','timeout'=>'90597366','version-github'=>'497af47b','version-noverify'=>'966937e9','clickhouse'=>'5bb80dfb','elastic'=>'f7017c4','firebird'=>'5499d1a','igdb'=>'170d083','imap'=>'ac143217','mongo'=>'c3b8f5a4','redis'=>'12f1a73b','simpledb'=>'79488f8b',);}function
__call($D,array$Jg){$ta=array();foreach($Jg
as$x=>$X)$ta[]=&$Jg[$x];$J=null;foreach($this->hooks[$D]as$ah){$Y=call_user_func_array(array($ah,$D),$ta);if($Y!==null){if(!self::$append[$D])return$Y;$J=$Y+(array)$J;}}return$J;}}abstract
class
Plugin{protected$translations=array();function
description(){return$this->lang('');}function
screenshot(){return"";}protected
function
lang($u,$Yf=null){$ta=func_get_args();$ta[0]=idx($this->translations[LANG],$u)?:$u;return
call_user_func_array('Adminer\lang_format',$ta);}}Adminer::$instance=(function_exists('adminer_object')?adminer_object():(is_dir("adminer-plugins")||file_exists("adminer-plugins.php")?new
Plugins(null):new
Adminer));SqlDriver::$drivers=array("server"=>"MySQL / MariaDB")+SqlDriver::$drivers;if(!defined('Adminer\DRIVER')){define('Adminer\DRIVER',"server");if(extension_loaded("mysqli")&&$_GET["ext"]!="pdo"){class
Db
extends
\MySQLi{static$instance;var$extension="MySQLi",$flavor='';function
__construct(){parent::init();}function
attach($N,$V,$Ug){mysqli_report(MYSQLI_REPORT_OFF);list($Ud,$dh)=host_port($N);$yi=adminer()->connectSsl();$Nj=($yi&&($yi['key']||$yi['cert']||$yi['ca']||isset($yi['verify'])));if($Nj)$this->ssl_set($yi['key'],$yi['cert'],$yi['ca'],'','');$J=@$this->real_connect(($N!=""?$Ud:ini_get("mysqli.default_host")),($N.$V!=""?$V:ini_get("mysqli.default_user")),($N.$V.$Ug!=""?$Ug:ini_get("mysqli.default_pw")),null,(is_numeric($dh)?intval($dh):ini_get("mysqli.default_port")),(is_numeric($dh)?null:$dh),($Nj?($yi['verify']!==false?MYSQLI_CLIENT_SSL:64):0));$this->options(MYSQLI_OPT_LOCAL_INFILE,0);return($J?'':$this->error);}function
set_charset($Ta){if(parent::set_charset($Ta))return
true;parent::set_charset('utf8');return$this->query("SET NAMES $Ta");}function
next_result(){return
self::more_results()&&parent::next_result();}function
quote($Q){return"'".$this->escape_string($Q)."'";}function
inTransaction(){return
false;}}}elseif(extension_loaded("mysql")&&!((ini_bool("sql.safe_mode")||ini_bool("mysql.allow_local_infile"))&&extension_loaded("pdo_mysql"))){class
Db
extends
SqlDb{private$link;function
attach($N,$V,$Ug){if(ini_bool("mysql.allow_local_infile"))return
sprintf('Disable %s or enable %s or %s extensions.',"'mysql.allow_local_infile'","MySQLi","PDO_MySQL");$this->link=@mysql_connect(($N!=""?$N:ini_get("mysql.default_host")),($N.$V!=""?$V:ini_get("mysql.default_user")),($N.$V.$Ug!=""?$Ug:ini_get("mysql.default_password")),true,131072);if(!$this->link)return
mysql_error();$this->server_info=mysql_get_server_info($this->link);return'';}function
set_charset($Ta){return
mysql_set_charset($Ta,$this->link)||mysql_set_charset('utf8',$this->link);}function
quote($Q){return"'".mysql_real_escape_string($Q,$this->link)."'";}function
select_db($Kb){return
mysql_select_db($Kb,$this->link);}function
query($H,$yj=false){$I=@($yj?mysql_unbuffered_query($H,$this->link):mysql_query($H,$this->link));$this->error="";if(!$I){$this->errno=mysql_errno($this->link);$this->error=mysql_error($this->link);return
false;}if($I===true){$this->affected_rows=mysql_affected_rows($this->link);$this->info=mysql_info($this->link);return
true;}return
new
Result($I);}}class
Result{var$num_rows;private$result;private$offset=0;function
__construct($I){$this->result=$I;$this->num_rows=mysql_num_rows($I);}function
fetch_assoc(){return
mysql_fetch_assoc($this->result);}function
fetch_row(){return
mysql_fetch_row($this->result);}function
fetch_field(){$J=mysql_fetch_field($this->result,$this->offset++);$J->orgtable=$J->table;$J->charsetnr=($J->blob?63:0);return$J;}}}elseif(extension_loaded("pdo_mysql")){class
Db
extends
PdoDb{var$extension="PDO_MySQL";function
attach($N,$V,$Ug){$sg=array(\PDO::MYSQL_ATTR_LOCAL_INFILE=>false);if(isset($_GET["select"]))$sg[\PDO::MYSQL_ATTR_MULTI_STATEMENTS]=false;$yi=adminer()->connectSsl();if($yi){if($yi['key'])$sg[\PDO::MYSQL_ATTR_SSL_KEY]=$yi['key'];if($yi['cert'])$sg[\PDO::MYSQL_ATTR_SSL_CERT]=$yi['cert'];if($yi['ca'])$sg[\PDO::MYSQL_ATTR_SSL_CA]=$yi['ca'];if(isset($yi['verify']))$sg[\PDO::MYSQL_ATTR_SSL_VERIFY_SERVER_CERT]=$yi['verify'];}list($Ud,$dh)=host_port($N);return$this->dsn("mysql:charset=utf8".($Ud!=""?";host=$Ud":'').($dh?(is_numeric($dh)?";port=":";unix_socket=").$dh:""),$V,$Ug,$sg);}function
set_charset($Ta){return$this->query("SET NAMES $Ta");}function
select_db($Kb){return$this->query("USE ".idf_escape($Kb));}function
query($H,$yj=false){$this->pdo->setAttribute(\PDO::MYSQL_ATTR_USE_BUFFERED_QUERY,!$yj);return
parent::query($H,$yj);}}}class
Driver
extends
SqlDriver{static$extensions=array("MySQLi","MySQL","PDO_MySQL");static$jush="sql";var$unsigned=array("unsigned","zerofill","unsigned zerofill");var$operators=array("=","<",">","<=",">=","!=","LIKE","LIKE %%","REGEXP","IN","FIND_IN_SET","IS NULL","NOT LIKE","NOT REGEXP","NOT IN","IS NOT NULL","SQL");var$functions=array("char_length","date","from_unixtime","lower","round","floor","ceil","sec_to_time","time_to_sec","upper");var$grouping=array("avg","count","count distinct","group_concat","max","min","sum");var$partitionBy=array("HASH","LINEAR HASH","KEY","LINEAR KEY","RANGE","LIST");static
function
connect($N,$V,$Ug){$e=parent::connect($N,$V,$Ug);if(is_string($e)){if(function_exists('iconv')&&!is_utf8($e)&&strlen($Th=iconv("windows-1252","utf-8//IGNORE",$e))>strlen($e))$e=$Th;return$e;}$e->set_charset(charset($e));$e->query("SET sql_quote_show_create = 1, autocommit = 1");$e->flavor=(preg_match('~MariaDB~',$e->server_info)?'maria':'mysql');add_driver(DRIVER,($e->flavor=='maria'?"MariaDB":"MySQL"));return$e;}function
__construct(Db$e){parent::__construct($e);$this->types=array('Numbers'=>array("tinyint"=>3,"smallint"=>5,"mediumint"=>8,"int"=>10,"bigint"=>20,"decimal"=>66,"float"=>12,"double"=>21),'Date and time'=>array("date"=>10,"datetime"=>19,"timestamp"=>19,"time"=>10,"year"=>4),'Strings'=>array("char"=>255,"varchar"=>65535,"tinytext"=>255,"text"=>65535,"mediumtext"=>16777215,"longtext"=>4294967295),'Lists'=>array("enum"=>65535,"set"=>64),'Binary'=>array("bit"=>20,"binary"=>255,"varbinary"=>65535,"tinyblob"=>255,"blob"=>65535,"mediumblob"=>16777215,"longblob"=>4294967295),'Geometry'=>array("geometry"=>0,"point"=>0,"linestring"=>0,"polygon"=>0,"multipoint"=>0,"multilinestring"=>0,"multipolygon"=>0,"geometrycollection"=>0),);$this->insertFunctions=array("char"=>"md5/sha1/password/encrypt/uuid","binary"=>"md5/sha1","date|time"=>"now",);$this->editFunctions=array(number_type()=>"+/-","date"=>"+ interval/- interval","time"=>"addtime/subtime","char|text"=>"concat",);if(min_version('5.7.8',10.2,$e))$this->types['Strings']["json"]=4294967295;if(min_version('',10.7,$e)){$this->types['Strings']["uuid"]=128;$this->insertFunctions['uuid']='uuid';}if(min_version('',10.5,$e)){$this->types['Network']["inet6"]=39;if(min_version('','10.10',$e))$this->types['Network']["inet4"]=15;}if(min_version(9,11.7,$e))$this->types['Numbers']["vector"]=16383;if(min_version(5.7,10.2,$e))$this->generated=array("STORED","VIRTUAL");}function
unconvertFunction(array$l){return(preg_match("~binary~",$l["type"])?"<code class='jush-sql'>UNHEX</code>":($l["type"]=="bit"?doc_link(array('sql'=>'bit-value-literals.html'),"<code>b''</code>"):($l["type"]=="vector"?"<code class='jush-sql'>".($this->conn->flavor=='maria'?"VEC_FromText":"STRING_TO_VECTOR")."</code>":(preg_match("~geometry|point|linestring|polygon~",$l["type"])?"<code class='jush-sql'>GeomFromText</code>":""))));}function
insert($R,array$O){return($O?parent::insert($R,$O):queries("INSERT INTO ".table($R)." ()\nVALUES ()"));}function
insertUpdate($R,array$L,array$mh){$d=array_keys(reset($L));$kh="INSERT INTO ".table($R)." (".implode(", ",$d).") VALUES\n";$Tj=array();foreach($d
as$x)$Tj[$x]="$x = VALUES($x)";$Ei="\nON DUPLICATE KEY UPDATE ".implode(", ",$Tj);$Tj=array();$y=0;foreach($L
as$O){$Y="(".implode(", ",$O).")";if($Tj&&(strlen($kh)+$y+strlen($Y)+strlen($Ei)>1e6)){if(!queries($kh.implode(",\n",$Tj).$Ei))return
false;$Tj=array();$y=0;}$Tj[]=$Y;$y+=strlen($Y)+2;}return
queries($kh.implode(",\n",$Tj).$Ei);}function
slowQuery($H,$aj){if(min_version('5.7.8','10.1.2')){if($this->conn->flavor=='maria')return"SET STATEMENT max_statement_time=$aj FOR $H";elseif(preg_match('~^(SELECT\b)(.+)~is',$H,$B))return"$B[1] /*+ MAX_EXECUTION_TIME(".($aj*1000).") */ $B[2]";}}function
convertSearch($u,array$X,array$l){return(preg_match('~char|text|enum|set~',$l["type"])&&!preg_match("~^utf8~",$l["collation"])&&preg_match('~[\x80-\xFF]~',$X['val'])?"CONVERT($u USING ".charset($this->conn).")":$u);}function
typeName(\stdClass$l){$xj=array("decimal","tinyint","smallint","int","float","double",7=>"timestamp","bigint","mediumint","date","time","datetime","year",15=>"varchar","bit",242=>"vector",245=>"json","decimal","enum","set","tinytext","mediumtext","longtext","text","varchar","char","geometry",);$J=idx($xj,$l->type,"");return
parent::typeName($l)?:($l->charsetnr==63?str_replace(array("text","varchar","char"),array("blob","varbinary","binary"),$J):$J);}function
quoteBinary($Th){return"X".q(bin2hex($Th));}function
warnings(){$I=$this->conn->query("SHOW WARNINGS");if($I&&$I->num_rows){ob_start();print_select_result($I);return
ob_get_clean();}}function
tableHelp($D,$Ce=false){$gf=($this->conn->flavor=='maria');if(information_schema(DB))return
strtolower(str_replace("_","-",DB)."-".($gf?"$D-table/":str_replace("_","-",$D)."-table.html"));if(DB=="sys")return($gf?"sys-schema/":strtolower("sys-".str_replace("_","-",preg_replace('~^x\$~','',$D)).".html"));if(DB=="mysql")return($gf?"mysql$D-table/":"system-schema.html");}function
partitionsInfo($R){$td="FROM information_schema.PARTITIONS WHERE TABLE_SCHEMA = ".q(DB)." AND TABLE_NAME = ".q($R);$I=$this->conn->query("SELECT PARTITION_METHOD, PARTITION_EXPRESSION, PARTITION_ORDINAL_POSITION $td ORDER BY PARTITION_ORDINAL_POSITION DESC LIMIT 1");$K=($I?$I->fetch_row():null);if(!$K)return
array();$J=array();list($J["partition_by"],$J["partition"],$J["partitions"])=$K;$Rg=get_key_vals("SELECT PARTITION_NAME, PARTITION_DESCRIPTION $td AND PARTITION_NAME != '' ORDER BY PARTITION_ORDINAL_POSITION");$J["partition_names"]=array_keys($Rg);$J["partition_values"]=array_values($Rg);return$J;}function
hasCStyleEscapes(){static$Oa;if($Oa===null){$wi=get_val("SHOW VARIABLES LIKE 'sql_mode'",1,$this->conn);$Oa=(strpos($wi,'NO_BACKSLASH_ESCAPES')===false);}return$Oa;}function
lineComment(){return"#|-- ";}function
engines(){$J=array();foreach(get_rows("SHOW ENGINES")as$K){if(preg_match("~YES|DEFAULT~",$K["Support"]))$J[]=$K["Engine"];}return$J;}function
indexAlgorithms(array$Ki){return(preg_match('~^(MEMORY|NDB)$~',$Ki["Engine"])?array("HASH","BTREE"):array());}}function
idf_escape($u){return"`".str_replace("`","``",$u)."`";}function
table($u){return
idf_escape($u);}function
get_databases($ld){$J=get_session("dbs");if($J===null){$H="SELECT SCHEMA_NAME FROM information_schema.SCHEMATA ORDER BY SCHEMA_NAME";$zi=microtime(true);$J=($ld?slow_query($H):get_vals($H));if(microtime(true)-$zi>0.1){restart_session();set_session("dbs",$J);stop_session();}}return$J;}function
limit($H,$Z,$z,$dg=0,$ei=" "){return" $H$Z".($z?$ei."LIMIT $z".($dg?" OFFSET $dg":""):"");}function
limit1($R,$H,$Z,$ei="\n"){return
limit($H,$Z,1,0,$ei);}function
db_collation($i,array$gb){$J=null;$g=get_val("SHOW CREATE DATABASE ".idf_escape($i),1);if(preg_match('~ COLLATE ([^ ]+)~',$g,$B))$J=$B[1];elseif(preg_match('~ CHARACTER SET ([^ ]+)~',$g,$B))$J=$gb[$B[1]][-1];return$J;}function
logged_user(){return
get_val("SELECT USER()");}function
tables_list(){return
get_key_vals("SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() ORDER BY TABLE_NAME");}function
count_tables(array$h){$J=array();foreach($h
as$i)$J[$i]=count(get_vals("SHOW TABLES IN ".idf_escape($i)));return$J;}function
table_status($D="",$Wc=false){$J=array();foreach(get_rows($Wc?"SELECT TABLE_NAME AS Name, ENGINE AS Engine, TABLE_COMMENT AS Comment FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() ".($D!=""?"AND TABLE_NAME = ".q($D):"ORDER BY Name"):"SHOW TABLE STATUS".($D!=""?" LIKE ".q(addcslashes($D,"%_\\")):""))as$K){if($K["Engine"]=="InnoDB")$K["Comment"]=preg_replace('~(?:(.+); )?InnoDB free: .*~','\1',$K["Comment"]);if(!isset($K["Engine"]))$K["Comment"]="";if($D!="")$K["Name"]=$D;$J[$K["Name"]]=$K;}return$J;}function
is_view(array$S){return$S["Engine"]===null;}function
fk_support(array$S){return
preg_match('~InnoDB|IBMDB2I'.(min_version(5.6)?'|NDB':'').'~i',$S["Engine"]);}function
parse_type($vd){preg_match('~^([^( ]+)(?:\((.+)\))?( unsigned)?( zerofill)?$~',$vd,$B);return
array($B[1],$B[2],ltrim($B[3].$B[4]));}function
fields($R){$gf=(connection()->flavor=='maria');$J=array();foreach(get_rows("SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ".q($R)." ORDER BY ORDINAL_POSITION")as$K){$l=$K["COLUMN_NAME"];$U=$K["COLUMN_TYPE"];$zd=$K["GENERATION_EXPRESSION"];$Tc=$K["EXTRA"];preg_match('~^(VIRTUAL|PERSISTENT|STORED)~',$Tc,$yd);list($wj,$y,$Dj)=parse_type($U);$j=$K["COLUMN_DEFAULT"];if($j!=""){$Be=preg_match('~text|json~',$wj);if(!$gf&&$Be)$j=preg_replace("~^(_\w+)?('.*')$~",'\2',stripslashes($j));if($gf||$Be){$j=($j=="NULL"?null:preg_replace_callback("~^'(.*)'$~",function($B){return
stripslashes(str_replace("''","'",$B[1]));},$j));}if(!$gf&&preg_match('~binary~',$wj)&&preg_match('~^0x(\w*)$~',$j,$B))$j=pack("H*",$B[1]);}$J[$l]=array("field"=>$l,"full_type"=>$U,"type"=>$wj,"length"=>$y,"unsigned"=>$Dj,"default"=>($yd?($gf?$zd:stripslashes($zd)):$j),"null"=>($K["IS_NULLABLE"]=="YES"),"auto_increment"=>($Tc=="auto_increment"),"on_update"=>(preg_match('~\bon update (\w+)~i',$Tc,$B)?$B[1]:""),"collation"=>$K["COLLATION_NAME"],"privileges"=>array_flip(explode(",","$K[PRIVILEGES],where,order")),"comment"=>$K["COLUMN_COMMENT"],"primary"=>($K["COLUMN_KEY"]=="PRI"),"generated"=>($yd[1]=="PERSISTENT"?"STORED":$yd[1]),);}return$J;}function
indexes($R,$f=null){$J=array();foreach(get_rows("SHOW INDEX FROM ".table($R),$f)as$K){$D=$K["Key_name"];$J[$D]["type"]=($D=="PRIMARY"?"PRIMARY":($K["Index_type"]=="FULLTEXT"?"FULLTEXT":($K["Non_unique"]?(preg_match('~^(SPATIAL|VECTOR)$~',$K["Index_type"])?$K["Index_type"]:"INDEX"):"UNIQUE")));$J[$D]["columns"][]=$K["Column_name"];$J[$D]["lengths"][]=($K["Index_type"]=="SPATIAL"?null:$K["Sub_part"]);$J[$D]["descs"][]=null;$J[$D]["algorithm"]=$K["Index_type"];}return$J;}function
foreign_keys($R){static$Wg='(?:`(?:[^`]|``)+`|"(?:[^"]|"")+")';$J=array();$zb=get_val("SHOW CREATE TABLE ".table($R),1);if($zb){preg_match_all("~CONSTRAINT ($Wg) FOREIGN KEY ?\\(((?:$Wg,? ?)+)\\) REFERENCES ($Wg)(?:\\.($Wg))? \\(((?:$Wg,? ?)+)\\)(?: ON DELETE (".driver()->onActions."))?(?: ON UPDATE (".driver()->onActions."))?~",$zb,$if,PREG_SET_ORDER);foreach($if
as$B){preg_match_all("~$Wg~",$B[2],$ri);preg_match_all("~$Wg~",$B[5],$Ti);$J[idf_unescape($B[1])]=array("db"=>idf_unescape($B[4]!=""?$B[3]:$B[4]),"table"=>idf_unescape($B[4]!=""?$B[4]:$B[3]),"source"=>array_map('Adminer\idf_unescape',$ri[0]),"target"=>array_map('Adminer\idf_unescape',$Ti[0]),"on_delete"=>($B[6]?:"RESTRICT"),"on_update"=>($B[7]?:"RESTRICT"),);}}return$J;}function
view($D){return
array("select"=>preg_replace('~^(?:[^`]|`[^`]*`)*\s+AS\s+~isU','',get_val("SHOW CREATE VIEW ".table($D),1)));}function
collations(){$J=array();foreach(get_rows("SHOW COLLATION")as$K){if($K["Default"])$J[$K["Charset"]][-1]=$K["Collation"];else$J[$K["Charset"]][]=$K["Collation"];}ksort($J);foreach($J
as$x=>$X)sort($J[$x]);return$J;}function
information_schema($i,$Vh=""){return($i=="information_schema")||(min_version(5.5)&&$i=="performance_schema");}function
error(){return
h(preg_replace('~^You have an error.*syntax to use~U',"Syntax error",connection()->error));}function
create_database($i,$fb){return
queries("CREATE DATABASE ".idf_escape($i).($fb?" COLLATE ".q($fb):""));}function
drop_databases(array$h){$J=apply_queries("DROP DATABASE",$h,'Adminer\idf_escape');restart_session();set_session("dbs",null);return$J;}function
rename_database($D,$fb){$J=false;if(create_database($D,$fb)){$T=array();$Yj=array();foreach(tables_list()as$R=>$U){if($U=='VIEW')$Yj[]=$R;else$T[]=$R;}$J=(!$T&&!$Yj)||move_tables($T,$Yj,$D);drop_databases($J?array(DB):array());}return$J;}function
auto_increment(){$_a=" PRIMARY KEY";if($_GET["create"]!=""&&$_POST["auto_increment_col"]){foreach(indexes($_GET["create"])as$v){if(in_array($_POST["fields"][$_POST["auto_increment_col"]]["orig"],$v["columns"],true)){$_a="";break;}if($v["type"]=="PRIMARY")$_a=" UNIQUE";}}return" AUTO_INCREMENT$_a";}function
alter_table($R,$D,array$m,array$nd,$jb,$zc,$fb,$za,$Qg){$pa=array();foreach($m
as$l){if($l[1]){$j=$l[1][3];if(preg_match('~ GENERATED~',$j)){$l[1][3]=(connection()->flavor=='maria'?"":$l[1][2]);$l[1][2]=$j;}$pa[]=($R!=""?($l[0]!=""?"CHANGE ".idf_escape($l[0]):"ADD"):" ")." ".implode($l[1]).($R!=""?$l[2]:"");}else$pa[]="DROP ".idf_escape($l[0]);}$pa=array_merge($pa,$nd);$P=($jb!==null?" COMMENT=".q($jb):"").($zc?" ENGINE=".q($zc):"").($fb?" COLLATE ".q($fb):"").($za!=""?" AUTO_INCREMENT=$za":"");if($Qg){$Rg=array();if($Qg["partition_by"]=='RANGE'||$Qg["partition_by"]=='LIST'){foreach($Qg["partition_names"]as$x=>$X){$Y=$Qg["partition_values"][$x];$Rg[]="\n  PARTITION ".idf_escape($X)." VALUES ".($Qg["partition_by"]=='RANGE'?"LESS THAN":"IN").($Y!=""?" ($Y)":" MAXVALUE");}}$P
.="\nPARTITION BY $Qg[partition_by]($Qg[partition])";if($Rg)$P
.=" (".implode(",",$Rg)."\n)";elseif($Qg["partitions"])$P
.=" PARTITIONS ".(+$Qg["partitions"]);}elseif($Qg===null)$P
.="\nREMOVE PARTITIONING";if($R=="")return
queries("CREATE TABLE ".table($D)." (\n".implode(",\n",$pa)."\n)$P");if($R!=$D)$pa[]="RENAME TO ".table($D);if($P)$pa[]=ltrim($P);return($pa?queries("ALTER TABLE ".table($R)."\n".implode(",\n",$pa)):true);}function
alter_indexes($R,$pa){$Ra=array();foreach($pa
as$X)$Ra[]=($X[2]=="DROP"?"\nDROP INDEX ".idf_escape($X[1]):"\nADD $X[0] ".($X[0]=="PRIMARY"?"KEY ":"").($X[1]!=""?idf_escape($X[1])." ":"")."(".implode(", ",$X[2]).")");return
queries("ALTER TABLE ".table($R).implode(",",$Ra));}function
truncate_tables(array$T){return
apply_queries("TRUNCATE TABLE",$T);}function
drop_views(array$Yj){return
queries("DROP VIEW ".implode(", ",array_map('Adminer\table',$Yj)));}function
drop_tables(array$T){return
queries("DROP TABLE ".implode(", ",array_map('Adminer\table',$T)));}function
move_tables(array$T,array$Yj,$Ti){$Hh=array();foreach($T
as$R)$Hh[]=table($R)." TO ".idf_escape($Ti).".".table($R);if(!$Hh||queries("RENAME TABLE ".implode(", ",$Hh))){$Tb=array();foreach($Yj
as$R)$Tb[table($R)]=view($R);connection()->select_db($Ti);$i=idf_escape(DB);foreach($Tb
as$D=>$Xj){if(!queries("CREATE VIEW $D AS ".str_replace(" $i."," ",$Xj["select"]))||!queries("DROP VIEW $i.$D"))return
false;}return
true;}return
false;}function
copy_tables(array$T,array$Yj,$Ti){queries("SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO'");foreach($T
as$R){$D=($Ti==DB?table("copy_$R"):idf_escape($Ti).".".table($R));if(($_POST["overwrite"]&&!queries("\nDROP TABLE IF EXISTS $D"))||!queries("CREATE TABLE $D LIKE ".table($R))||!queries("INSERT INTO $D SELECT * FROM ".table($R)))return
false;foreach(get_rows("SHOW TRIGGERS LIKE ".q(addcslashes($R,"%_\\")))as$K){$pj=$K["Trigger"];if(!queries("CREATE TRIGGER ".($Ti==DB?idf_escape("copy_$pj"):idf_escape($Ti).".".idf_escape($pj))." $K[Timing] $K[Event] ON $D FOR EACH ROW\n$K[Statement];"))return
false;}}foreach($Yj
as$R){$D=($Ti==DB?table("copy_$R"):idf_escape($Ti).".".table($R));$Xj=view($R);if(($_POST["overwrite"]&&!queries("DROP VIEW IF EXISTS $D"))||!queries("CREATE VIEW $D AS $Xj[select]"))return
false;}return
true;}function
trigger($D,$R){if($D=="")return
array();$L=get_rows("SHOW TRIGGERS WHERE `Trigger` = ".q($D));return
reset($L);}function
triggers($R){$J=array();foreach(get_rows("SHOW TRIGGERS LIKE ".q(addcslashes($R,"%_\\")))as$K)$J[$K["Trigger"]]=array($K["Timing"],$K["Event"]);return$J;}function
trigger_options(){return
array("Timing"=>array("BEFORE","AFTER"),"Event"=>array("INSERT","UPDATE","DELETE"),"Type"=>array("FOR EACH ROW"),);}function
routine($D,$U){$L=get_rows("SELECT PARAMETER_NAME, DTD_IDENTIFIER, PARAMETER_MODE, CHARACTER_SET_NAME
FROM information_schema.PARAMETERS
WHERE SPECIFIC_SCHEMA = DATABASE() AND ROUTINE_TYPE = '$U' AND SPECIFIC_NAME = ".q($D)."
ORDER BY ORDINAL_POSITION");$m=array();foreach($L
as$K){$vd=$K["DTD_IDENTIFIER"];list($wj,$y,$Dj)=parse_type($vd);$m[]=array("field"=>$K["PARAMETER_NAME"],"type"=>$wj,"length"=>$y,"unsigned"=>$Dj,"null"=>true,"full_type"=>$vd,"inout"=>($U=="FUNCTION"?"":$K["PARAMETER_MODE"]),"collation"=>$K["CHARACTER_SET_NAME"],);}$J=connection()->query("SELECT
	ROUTINE_COMMENT comment,
	CONCAT(IF(IS_DETERMINISTIC = 'YES', 'DETERMINISTIC\\n', ''), IF(SQL_DATA_ACCESS != 'CONTAINS SQL', CONCAT(SQL_DATA_ACCESS, '\\n'), ''), ROUTINE_DEFINITION) definition,
	'SQL' language
FROM information_schema.ROUTINES
WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_TYPE = '$U' AND ROUTINE_NAME = ".q($D))->fetch_assoc();if($m&&$m[0]['field']=='')$J['returns']=array_shift($m);$J['fields']=$m;return$J;}function
routines(){return
get_rows("SELECT SPECIFIC_NAME, ROUTINE_NAME, ROUTINE_TYPE, DTD_IDENTIFIER FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = DATABASE()");}function
routine_languages(){return
array();}function
routine_id($D,array$K){return
idf_escape($D);}function
last_id($I){return
get_val("SELECT LAST_INSERT_ID()");}function
explain(Db$e,$H){return$e->query("EXPLAIN ".(min_version(5.7)?"":"PARTITIONS ").$H);}function
found_rows(array$S,array$Z){return($Z||$S["Engine"]!="InnoDB"?null:$S["Rows"]);}function
create_sql($R,$za,$Ci){$J=get_val("SHOW CREATE TABLE ".table($R),1);if(!$za)$J=preg_replace('~(\n\)[^\n]*?) AUTO_INCREMENT=\d+~','\1',$J);return$J;}function
truncate_sql($R){return"TRUNCATE ".table($R);}function
use_sql($Kb,$Ci=""){$D=idf_escape($Kb);$J="";if(preg_match('~CREATE~',$Ci)&&($g=get_val("SHOW CREATE DATABASE $D",1))){set_utf8mb4($g);if($Ci=="DROP+CREATE")$J="DROP DATABASE IF EXISTS $D;\n";$J
.="$g;\n";}return$J."USE $D";}function
trigger_sql($R){$J="";foreach(get_rows("SHOW TRIGGERS LIKE ".q(addcslashes($R,"%_\\")),null,"-- ")as$K)$J
.="\nCREATE TRIGGER ".idf_escape($K["Trigger"])." $K[Timing] $K[Event] ON ".table($K["Table"])." FOR EACH ROW\n$K[Statement];;\n";return$J;}function
show_variables(){return
get_rows("SHOW VARIABLES");}function
show_status(){return
get_rows("SHOW STATUS");}function
process_list(){return
get_rows("SHOW FULL PROCESSLIST");}function
convert_field(array$l){if(preg_match("~binary~",$l["type"]))return"HEX(".idf_escape($l["field"]).")";if($l["type"]=="bit")return"BIN(".idf_escape($l["field"])." + 0)";if($l["type"]=="vector")return(connection()->flavor=='maria'?"VEC_ToText":"VECTOR_TO_STRING")."(".idf_escape($l["field"]).")";if(preg_match("~geometry|point|linestring|polygon~",$l["type"]))return(min_version(8)?"ST_":"")."AsWKT(".idf_escape($l["field"]).")";}function
unconvert_field(array$l,$J){if(preg_match("~binary~",$l["type"]))$J="UNHEX($J)";if($l["type"]=="bit")$J="CONVERT(b$J, UNSIGNED)";if($l["type"]=="vector")$J=(connection()->flavor=='maria'?"VEC_FromText":"STRING_TO_VECTOR")."($J)";if(preg_match("~geometry|point|linestring|polygon~",$l["type"])){$kh=(min_version(8)?"ST_":"");$J=$kh."GeomFromText($J, $kh"."SRID($l[field]))";}return$J;}function
support($Xc){return
preg_match('~^(comment|columns|copy|database|drop_col|dump|event|indexes|kill|privileges|move_col|procedure|processlist|routine|sql|status|table|trigger|variables|view'.(min_version(8)?'|descidx':'').(min_version('8.0.16','10.2.1')?'|check':'').(min_version(8,99)?'|fast_status':'').')$~',$Xc);}function
kill_process($t){return
queries("KILL ".number($t));}function
connection_id(){return"SELECT CONNECTION_ID()";}function
max_connections(){return
get_val("SELECT @@max_connections");}function
types($Sc=false){return
array();}function
type_values($t){return"";}function
type_definition($t){return
array("kind"=>"","definition"=>"");}function
schemas(){return
array();}function
get_schema(){return"";}function
set_schema($Vh,$f=null){return
true;}}define('Adminer\JUSH',Driver::$jush);define('Adminer\SERVER',"".$_GET[DRIVER]);define('Adminer\DB',"$_GET[db]");define('Adminer\ME',preg_replace('~\?.*~','',relative_uri()).'?'.(sid()?SID.'&':'').($_GET["ext"]?"ext=".url_escape($_GET["ext"]).'&':'').(isset($_GET[DRIVER])?DRIVER."=".url_escape(SERVER).'&':'').(isset($_GET["username"])?"username=".url_escape($_GET["username"]).'&':'').(DB!=""?'db='.url_escape(DB).'&'.(isset($_GET["ns"])?"ns=".url_escape($_GET["ns"])."&":""):''));function
page_header($cj,$k="",$Ma=array(),$dj=""){page_headers();if(is_ajax()&&$k){page_messages($k);exit;}if(!ob_get_level())ob_start('ob_gzhandler',4096);$ej=$cj.($dj!=""?": $dj":"");$fj=strip_tags($ej.(SERVER!=""&&SERVER!="localhost"?h(" - ".SERVER):"")." - ".adminer()->name());echo'<!DOCTYPE html>
<html lang=\'en\' dir=\'ltr\' class=\'ltr nojs\'>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta name="robots" content="noindex">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>',$fj,'</title>
<link rel="stylesheet" href="',h(preg_replace("~\\?.*~","",ME)."?file=default.css&version=6.0.0"),'">
';$Cb=adminer()->css();if(is_int(key($Cb)))$Cb=array_fill_keys($Cb,'light');$Jd=in_array('light',$Cb)||in_array('',$Cb);$Hd=in_array('dark',$Cb)||in_array('',$Cb);$Gb=($Jd?($Hd?null:false):($Hd?:null));$wf=" media='(prefers-color-scheme: dark)'";if($Gb!==false)echo"<link rel='stylesheet'".($Gb?"":$wf)." href='".h(preg_replace("~\\?.*~","",ME)."?file=dark.css&version=6.0.0")."'>\n";echo"<meta name='color-scheme' content='".($Gb===null?"light dark":($Gb?"dark":"light"))."'>\n",script_src(preg_replace("~\\?.*~","",ME)."?file=functions.js&version=6.0.0");if(adminer()->head($Gb))echo"<link rel='icon' href='data:image/gif;base64,"."R0lGODlhEAAQAJEAAAQCBPz+/PwCBAROZCH5BAEAAAAALAAAAAAQABAAAAI2hI+pGO1rmghihiUdvUBnZ3XBQA7f05mOak1RWXrNq5nQWHMKvuoJ37BhVEEfYxQzHjWQ5qIAADs='>\n","<link rel='apple-touch-icon' href='".h(preg_replace("~\\?.*~","",ME)."?file=logo.png&version=6.0.0")."'>\n";foreach($Cb
as$Ij=>$If){$b=($If=='dark'&&!$Gb?$wf:($If=='light'&&$Hd?" media='(prefers-color-scheme: light)'":""));echo"<link rel='stylesheet'$b href='".h($Ij)."'>\n";}echo"\n<body class='";adminer()->bodyClass();echo"'>\n",script((isset($_COOKIE["adminer_version"])||!adminer()->verifyVersion()?"":"onload = partial(verifyVersion, '".VERSION."');\n")."
const offlineMessage = '".js_escape('You are offline.')."';
const thousandsSeparator = '".js_escape(',')."';
const urlSeparators = '".js_escape(ini_get("arg_separator.input"))."';"),"<div id='help' class='jush-".JUSH." jsonly hidden'".on('mouseover','helpKeep').on('mouseout','helpMouseout')."></div>\n","<div id='content'>\n","<span id='menuopen' class='jsonly'".on('click','menuToggle')."><button title='".'Menu'."' class='icon icon-move' aria-expanded='false'></button></span>\n";if($Ma!==null){$_=substr(preg_replace('~\b(username|db|ns)=[^&]*&~','',ME),0,-1);echo'<p id="breadcrumb"><a href="'.h($_?:".").'">'.get_driver(DRIVER).'</a> » ';$_=substr(preg_replace('~\b(db|ns)=[^&]*&~','',ME),0,-1);$N=adminer()->serverName(SERVER);$N=($N!=""?$N:'Server');if($Ma===false)echo"$N\n";else{echo"<a href='".h($_)."' accesskey='1' title='Alt+Shift+1'>$N</a> » ";if($_GET["ns"]!=""||(DB!=""&&is_array($Ma)))echo'<a href="'.h($_."&db=".url_escape(DB).(support("scheme")?"&ns=":"")).'">'.h(DB).'</a> » ';if(is_array($Ma)){if($_GET["ns"]!="")echo'<a href="'.h(substr(ME,0,-1)).'">'.h($_GET["ns"]).'</a> » ';foreach($Ma
as$x=>$X){$Vb=(is_array($X)?$X[1]:h($X));if($Vb!="")echo"<a href='".h(ME."$x=").url_escape(is_array($X)?$X[0]:$X)."'>$Vb</a> » ";}}echo"$cj\n";}}echo"<h2>$ej</h2>\n","<div id='ajaxstatus' role='status' class='jsonly'></div>\n";restart_session();page_messages($k);$h=&get_session("dbs");if(DB!=""&&$h&&!in_array(DB,$h,true))$h=null;stop_session();define('Adminer\PAGE_HEADER',1);ob_flush();flush();}function
page_headers(){header("Content-Type: text/html; charset=utf-8");header("Cache-Control: no-cache");header("X-Frame-Options: deny");header("X-XSS-Protection: 0");header("X-Content-Type-Options: nosniff");header("Referrer-Policy: origin-when-cross-origin");foreach(adminer()->csp(csp())as$Bb){$Md=array();foreach($Bb
as$x=>$X)$Md[]="$x $X";header("Content-Security-Policy: ".implode("; ",$Md));}adminer()->headers();}function
csp(){return
array(array("script-src"=>"'self' 'unsafe-inline' 'nonce-".get_nonce()."' 'strict-dynamic'","connect-src"=>"'self' https://www.adminer.org","frame-src"=>"https://www.adminer.org","object-src"=>"'none'","base-uri"=>"'none'","form-action"=>"'self'",),);}function
design_checksums(){$Oj=array();foreach(array_keys(adminer()->css())as$Ij)$Oj[preg_replace('~\?.*~','',$Ij)]=true;$J=array();foreach(array("adminer.css","adminer-dark.css")as$n){if($Oj[$n]&&file_exists($n)){preg_match('~^/\* Adminer design ([-\w]+) \*/~',file_get_contents($n),$B);$J[$n]=array((string)$B[1],Plugins::checksum($n));}}return$J;}function
official_design_checksums(){return
array('adminer-border/adminer-dark.css'=>'b2527e3','adminer-border/adminer.css'=>'430977ad','adminer-dark/adminer-dark.css'=>'a26bcd7b','brade/adminer.css'=>'be4161f0','bueltge/adminer.css'=>'1a8f00b4','dracula/adminer-dark.css'=>'cfaf61dd','esterka/adminer.css'=>'1f805f36','flat/adminer.css'=>'49a61af9','galkaev/adminer-dark.css'=>'16c46f94','haeckel/adminer.css'=>'147a3565','hever/adminer.css'=>'78b8cd43','konya/adminer.css'=>'3cc606c5','lavender-light/adminer.css'=>'bf03f5d7','lucas-sandery/adminer.css'=>'6596353','mancave/adminer-dark.css'=>'e1ac813d','mvt/adminer.css'=>'ebd3afdc','nette/adminer.css'=>'5ab360e7','ng9/adminer.css'=>'488583cf','nicu/adminer.css'=>'ecb9bd1e','pappu687/adminer.css'=>'b58d128c','paranoiq/adminer.css'=>'64d27e5','pepa-linha/adminer.css'=>'baf25f0','pokorny/adminer.css'=>'ee9eea6d','price/adminer.css'=>'b3c939b2','rmsoft/adminer.css'=>'391d54ad','rmsoft_blue-dark/adminer.css'=>'17714d77','rmsoft_blue/adminer.css'=>'c0f192ea','win98/adminer.css'=>'e82d63c3',);}function
version_iframe(){return(isset($_COOKIE["adminer_version"])||!adminer()->verifyVersion()?"":"<noscript><iframe sandbox src='https://www.adminer.org/version/?current=".VERSION."&amp;noscript=1'></iframe></noscript>");}function
get_nonce(){static$Vf;if(!$Vf)$Vf=base64_encode(rand_string());return$Vf;}function
page_messages($k){$Hj=preg_replace('~^[^?]*~','',$_SERVER["REQUEST_URI"]);$Bf=idx($_SESSION["messages"],$Hj);if($Bf){echo"<div class='message'>".implode("</div>\n<div class='message'>",$Bf)."</div>".script("messagesPrint();");unset($_SESSION["messages"][$Hj]);}if($k)echo"<div class='error'>$k</div>\n";if(adminer()->error)echo"<div class='error'>".adminer()->error."</div>\n";}function
page_footer($Hf=""){echo"</div>\n\n<div id='foot' class='foot'>\n<div id='menu'>\n";adminer()->navigation($Hf);echo"</div>\n";if($Hf!="auth")echo'<form action="" method="post">
<p class="logout">
<span title="Username">',h($_GET["username"])."\n",'</span>
<input type=\'submit\' name=\'logout\' value=\'Logout\' id=\'logout\'>
',input_token(),'</form>
';echo"</div>\n\n",script("setupSubmitHighlight(document);");}function
int32($Mf){while($Mf>=2147483648)$Mf-=4294967296;while($Mf<=-2147483649)$Mf+=4294967296;return(int)$Mf;}function
long2str(array$W,$ak){$Th='';foreach($W
as$X)$Th
.=pack('V',$X);if($ak)return
substr($Th,0,end($W));return$Th;}function
str2long($Th,$ak){$W=array_values(unpack('V*',str_pad($Th,4*ceil(strlen($Th)/4),"\0")));if($ak)$W[]=strlen($Th);return$W;}function
xxtea_mx($ik,$hk,$Fi,$He){return
int32((($ik>>5&0x7FFFFFF)^$hk<<2)+(($hk>>3&0x1FFFFFFF)^$ik<<4))^int32(($Fi^$hk)+($He^$ik));}function
encrypt_string($Ai,$x){if($Ai=="")return"";$x=array_values(unpack("V*",pack("H*",md5($x))));$W=str2long($Ai,true);$Mf=count($W)-1;$ik=$W[$Mf];$hk=$W[0];$th=floor(6+52/($Mf+1));$Fi=0;while($th-->0){$Fi=int32($Fi+0x9E3779B9);$rc=$Fi>>2&3;for($Gg=0;$Gg<$Mf;$Gg++){$hk=$W[$Gg+1];$Lf=xxtea_mx($ik,$hk,$Fi,$x[$Gg&3^$rc]);$ik=int32($W[$Gg]+$Lf);$W[$Gg]=$ik;}$hk=$W[0];$Lf=xxtea_mx($ik,$hk,$Fi,$x[$Gg&3^$rc]);$ik=int32($W[$Mf]+$Lf);$W[$Mf]=$ik;}return
long2str($W,false);}function
decrypt_string($Ai,$x){if($Ai=="")return"";if(!$x)return
false;$x=array_values(unpack("V*",pack("H*",md5($x))));$W=str2long($Ai,false);$Mf=count($W)-1;$ik=$W[$Mf];$hk=$W[0];$th=floor(6+52/($Mf+1));$Fi=int32($th*0x9E3779B9);while($Fi){$rc=$Fi>>2&3;for($Gg=$Mf;$Gg>0;$Gg--){$ik=$W[$Gg-1];$Lf=xxtea_mx($ik,$hk,$Fi,$x[$Gg&3^$rc]);$hk=int32($W[$Gg]-$Lf);$W[$Gg]=$hk;}$ik=$W[$Mf];$Lf=xxtea_mx($ik,$hk,$Fi,$x[$Gg&3^$rc]);$hk=int32($W[0]-$Lf);$W[0]=$hk;$Fi=int32($Fi-0x9E3779B9);}return
long2str($W,true);}$Yg=array();if($_COOKIE["adminer_permanent"]){foreach(explode(" ",$_COOKIE["adminer_permanent"])as$X){list($x)=explode(":",$X);$Yg[$x]=$X;}}function
add_invalid_login(){$Fa=get_temp_dir()."/adminer-invalid";foreach(glob("$Fa*")?:array($Fa)as$n){$p=file_open_lock($n);if($p)break;}if(!$p)$p=file_open_lock("$Fa-".rand_string());if(!$p)return;$we=json_decode(stream_get_contents($p),true);$Zi=time();if($we){foreach($we
as$xe=>$X){if($X[0]<$Zi)unset($we[$xe]);}}$ve=&$we[adminer()->bruteForceKey()];if(!$ve)$ve=array($Zi+30*60,0);$ve[1]++;file_write_unlock($p,json_encode($we));}function
check_invalid_login(array&$Yg){$we=array();foreach(glob(get_temp_dir()."/adminer-invalid*")as$n){$p=file_open_lock($n);if($p){$we=json_decode(stream_get_contents($p),true);file_unlock($p);break;}}$x=adminer()->bruteForceKey();$ve=idx($we,$x,array());$Uf=($ve[1]>29?$ve[0]-time():0);if($Uf>0){$k=lang_format(array('Too many unsuccessful logins, try again in %d minute.','Too many unsuccessful logins, try again in %d minutes.'),ceil($Uf/60));if($_SERVER["HTTP_X_FORWARDED_FOR"]!=""&&$x==$_SERVER["REMOTE_ADDR"])$k
.='<br>'.sprintf('Use the %s <a%s>plugin</a> if Adminer runs behind a reverse proxy.','<b>login-reverse-proxy</b>'," href='https://www.adminer.org/plugins/?version=".VERSION."'".target_blank());auth_error($k,$Yg);}}function
password_required(){static$J;if($J===null){$J=(bool)get_session("password_required");if(!$J){$Ab=adminer()->credentials();$J=!is_object(Driver::connect($Ab[0],$Ab[1],""));if($J)set_session("password_required",true);}}return$J;}$ya=$_POST["auth"];if($ya){session_regenerate_id();$Vj=$ya["driver"];$N=$ya["server"];$V=$ya["username"];$Ug=(string)$ya["password"];$i=$ya["db"];set_password($Vj,$N,$V,$Ug);$_SESSION["db"][$Vj][$N][$V][$i]=true;if($ya["permanent"]){$x=implode("-",array_map('base64_encode',array($Vj,$N,$V,$i)));$oh=adminer()->permanentLogin(true);$Yg[$x]="$x:".base64_encode($oh?encrypt_string($Ug,$oh):"");cookie("adminer_permanent",implode(" ",$Yg));}if(count($_POST)==1||DRIVER!=$Vj||SERVER!=$N||$_GET["username"]!==$V||DB!=$i)redirect(auth_url($Vj,$N,$V,$i));}elseif($_POST["logout"]&&(!$_SESSION["token"]||verify_token())){foreach(array("pwds","db","dbs","queries")as$x)set_session($x,null);unset_permanent($Yg);redirect(substr(preg_replace('~\b(username|db|ns)=[^&]*&~','',ME),0,-1),'Logout successful.'.' '.'Thanks for using Adminer, consider <a href="https://www.adminer.org/en/donation/">donating</a>.');}elseif($Yg&&!$_SESSION["pwds"]){session_regenerate_id();$oh=adminer()->permanentLogin();foreach($Yg
as$x=>$X){list(,$ab)=explode(":",$X);list($Vj,$N,$V,$i)=array_map('base64_decode',explode("-",$x));set_password($Vj,$N,$V,decrypt_string(base64_decode($ab),$oh));$_SESSION["db"][$Vj][$N][$V][$i]=true;}}function
unset_permanent(array&$Yg){foreach($Yg
as$x=>$X){list($Vj,$N,$V,$i)=array_map('base64_decode',explode("-",$x));if($Vj==DRIVER&&$N==SERVER&&$V==$_GET["username"]&&$i==DB)unset($Yg[$x]);}cookie("adminer_permanent",implode(" ",$Yg));}function
auth_error($k,array&$Yg){$hi=session_name();if(isset($_GET["username"])){header("HTTP/1.1 403 Forbidden");if(($_COOKIE[$hi]||$_GET[$hi])&&!$_SESSION["token"])$k='Session expired, please login again.';else{restart_session();add_invalid_login();$Ug=get_password();if($Ug!==null){if($Ug===false)$k
.=($k?'<br>':'').sprintf('Master password expired. <a href="https://www.adminer.org/en/extension/"%s>Implement</a> %s method to make it permanent.',target_blank(),'<code>permanentLogin()</code>');set_password(DRIVER,SERVER,$_GET["username"],null);}unset_permanent($Yg);}}if(!$_COOKIE[$hi]&&$_GET[$hi]&&ini_bool("session.use_only_cookies"))$k='Session support must be enabled.';$Jg=session_get_cookie_params();cookie("adminer_key",($_COOKIE["adminer_key"]?:rand_string()),$Jg["lifetime"]);if(!$_SESSION["token"])$_SESSION["token"]=rand(1,1e6);page_header('Login',$k,null);echo"<form action='' method='post'>\n","<div>";if(hidden_fields($_POST,array("auth")))echo"<p class='message'>".'The action will be performed after successful login with the same credentials.'."\n";echo"</div>\n";adminer()->loginForm();echo"</form>\n";page_footer("auth");exit;}if(isset($_GET["username"])&&!class_exists('Adminer\Db')){unset($_SESSION["pwds"][DRIVER]);unset_permanent($Yg);page_header('No extension',sprintf('None of the supported PHP extensions (%s) are available.',implode(", ",Driver::$extensions)),false);page_footer("auth");exit;}$e='';if(isset($_GET["username"])&&is_string(get_password())){list($Ud,$dh)=host_port(SERVER);if(preg_match('~[^-\w.:/]~',$Ud.$dh))auth_error('Invalid server.',$Yg);if(preg_match('~^-?\d+~',$dh,$B)&&($B[0]<1024||$B[0]>65535))auth_error('Connecting to privileged ports is not allowed.',$Yg);check_invalid_login($Yg);$Ab=adminer()->credentials();$e=Driver::connect($Ab[0],$Ab[1],$Ab[2]);if(is_object($e)){Db::$instance=$e;Driver::$instance=new
Driver($e);if($e->flavor)save_settings(array("vendor-".DRIVER."-".SERVER=>get_driver(DRIVER)));}}$ef=null;if(!is_object($e)||($ef=adminer()->login($_GET["username"],get_password()))!==true){$k=(is_string($e)?nl_br(h($e)):(is_string($ef)?$ef:'Invalid credentials.')).(preg_match('~^ | $~',get_password())?'<br>'.'There is a space in the input password which might be the cause.':'');auth_error($k,$Yg);}if($_POST["logout"]&&$_SESSION["token"]&&!verify_token()){page_header('Logout','Invalid CSRF token. Send the form again.');page_footer("db");exit;}if(!$_SESSION["token"])$_SESSION["token"]=rand(1,1e6);stop_session(true);if($ya&&$_POST["token"])$_POST["token"]=get_token();$k='';if($_POST){if(!verify_token())$k='Invalid CSRF token. Send the form again.'.' '.'If you did not send this request from Adminer then close this page.';}elseif($_SERVER["REQUEST_METHOD"]=="POST"){$k=sprintf('Too big POST data. Reduce the data or increase the %s configuration directive.',"<b>post_max_size</b>'");if(isset($_GET["sql"]))$k
.=' '.'You can upload a big SQL file via FTP and import it from server.';}function
print_select_result($I,$f=null,array$xg=array(),&$z=0){$af=array();$w=array();$d=array();$Ka=array();$xj=array();$J=array();for($s=0;(!$z||$s<$z)&&($K=$I->fetch_row());$s++){if(!$s){echo"<div class='scrollable'>\n","<table class='nowrap odds'>\n","<thead><tr>";for($De=0;$De<count($K);$De++){$l=$I->fetch_field();$D=$l->name;$wg=(isset($l->orgtable)?$l->orgtable:"");$vg=(isset($l->orgname)?$l->orgname:$D);if($xg&&JUSH=="sql")$af[$De]=($D=="table"?"table=":($D=="possible_keys"?"indexes=":null));elseif($wg!=""){if(isset($l->table))$J[$l->table]=$wg;if(!isset($w[$wg])){$w[$wg]=array();foreach(indexes($wg,$f)as$v){if($v["type"]=="PRIMARY"){$w[$wg]=array_flip($v["columns"]);break;}}$d[$wg]=$w[$wg];}if(isset($d[$wg][$vg])){unset($d[$wg][$vg]);$w[$wg][$vg]=$De;$af[$De]=$wg;}}if($l->charsetnr==63)$Ka[$De]=true;$xj[$De]=$l->type;echo"<th title='".h(trim(($wg!=""?"$wg.$vg":($l->name!=$vg?$vg:""))." ".driver()->typeName($l)))."'>".h($D).($xg?doc_link(array('sql'=>"explain-output.html#explain_".strtolower($D),'mariadb'=>"explain/#the-columns-in-explain-select",)):"");}echo"<tbody>\n";}echo"<tr>";foreach($K
as$x=>$X){$_="";if(isset($af[$x])&&!$d[$af[$x]]){if($xg&&JUSH=="sql"){$R=$K[array_search("table=",$af)];$_=ME.$af[$x].url_escape($xg[$R]!=""?$xg[$R]:$R);}else{$_=ME."edit=".url_escape($af[$x]);foreach($w[$af[$x]]as$db=>$De){if($K[$De]===null){$_="";break;}$_
.="&where[".url_escape(bracket_escape($db))."]=".url_escape($K[$De]);}}}$l=array('type'=>($Ka[$x]?'blob':($xj[$x]==254?'char':'')),);$X=select_value($X,$_,$l,null);echo"<td".($xj[$x]<=9||$xj[$x]==246?" class='number'":"").">$X";}}$z=$s;echo($s?"</table>\n</div>":"<p class='message'>".'No rows.')."\n";return$J;}function
referencable_primary($ci){$J=array();foreach(table_status('',true)as$Li=>$R){if($Li!=$ci&&fk_support($R)){foreach(fields($Li)as$l){if($l["primary"]){if($J[$Li]){unset($J[$Li]);break;}$J[$Li]=$l;}}}}return$J;}function
textarea($D,$Y,$L=10,$hb=80){echo"<textarea name='".h($D)."' rows='$L' cols='$hb' class='sqlarea jush-".JUSH."' spellcheck='false' wrap='off'>";if(is_array($Y)){foreach($Y
as$X)echo
h($X[0])."\n\n\n";}else
echo
h($Y);echo"</textarea>";}function
select_input($b,array$sg,$Y="",$Zg=""){if($sg&&$Y!=""&&!isset($sg[$Y]))$sg=array($Y=>$Y)+$sg;$Si=($sg?"select":"input");return"<$Si$b".($sg?"><option value=''>$Zg".optionlist($sg,$Y,true)."</select>":" size='10' value='".h($Y)."' placeholder='$Zg'>");}function
json_row($x,$X=null,$Gc=true){static$id=true;if($id)echo"{";if($x!=""){echo($id?"":",")."\n\t\"".addcslashes($x,"\r\n\t\"\\/").'": '.($X!==null?($Gc?'"'.addcslashes($X,"\r\n\"\\/").'"':$X):'null');$id=false;}else{echo"\n}\n";$id=true;}}function
edit_type($x,array$l,array$gb,array$pd=array(),array$Uc=array()){$U=(string)$l["type"];echo"<td><select name='".h($x)."[type]' class='type' aria-labelledby='label-type'".on_help_value().">";if($U&&!array_key_exists($U,driver()->types())&&!isset($pd[$U])&&!in_array($U,$Uc))$Uc[]=$U;$Bi=driver()->structuredTypes();if($pd)$Bi['Foreign keys']=$pd;echo
optionlist(array_merge($Uc,$Bi),$U),"</select><td>","<input name='".h($x)."[length]' value='".h($l["length"])."' size='3'".(!$l["length"]&&preg_match('~var(char|binary)$~',$U)?" class='required'":"")." aria-labelledby='label-length'>","<td class='options'>",($gb?"<input list='collations' name='".h($x)."[collation]'".option_types($U,'(char|text|enum|set)$')." value='".h($l["collation"])."' placeholder='(".'collation'.")'>":''),(driver()->unsigned?"<select name='".h($x)."[unsigned]'".option_types($U,'^$|'.number_type()).'><option>'.optionlist(driver()->unsigned,$l["unsigned"]).'</select>':''),(isset($l['on_update'])?"<select name='".h($x)."[on_update]'".option_types($U,'timestamp|datetime').'>'.optionlist(array(""=>"(".'ON UPDATE'.")","CURRENT_TIMESTAMP"),(preg_match('~^CURRENT_TIMESTAMP~i',$l["on_update"])?"CURRENT_TIMESTAMP":$l["on_update"])).'</select>':''),($pd?"<select name='".h($x)."[on_delete]'".option_types($U,'`')."><option value=''>(".'ON DELETE'.")".optionlist(explode("|",driver()->onActions),$l["on_delete"])."</select> ":" ");}function
option_types($U,$xj){return" data-types='".h($xj)."'".(preg_match("~$xj~",$U)?"":" class='hidden'");}function
process_length($y){$Bc=driver()->enumLength;return(preg_match("~^\\s*\\(?\\s*$Bc(?:\\s*,\\s*$Bc)*+\\s*\\)?\\s*\$~",$y)&&preg_match_all("~$Bc~",$y,$if)?"(".implode(",",$if[0]).")":preg_replace('~^[0-9].*~','(\0)',preg_replace('~[^-0-9,+()[\]]~','',$y)));}function
process_type(array$l,$eb="COLLATE"){return" $l[type]".process_length($l["length"]).(preg_match(number_type(),$l["type"])&&in_array($l["unsigned"],driver()->unsigned)?" $l[unsigned]":"").(preg_match('~char|text|enum|set~',$l["type"])&&$l["collation"]?" $eb ".(JUSH=="mssql"?$l["collation"]:q($l["collation"])):"");}function
process_field(array$l,array$vj){if($l["on_update"])$l["on_update"]=str_ireplace("current_timestamp()","CURRENT_TIMESTAMP",$l["on_update"]);return
array(idf_escape(trim($l["field"])),process_type($vj),($l["null"]?" NULL":" NOT NULL"),default_value($l),(preg_match('~timestamp|datetime~',$l["type"])&&$l["on_update"]?" ON UPDATE $l[on_update]":""),(support("comment")&&$l["comment"]!=""?" COMMENT ".q($l["comment"]):""),($l["auto_increment"]?auto_increment():null),);}function
default_value(array$l){if($l["default"]===null)return"";$j=str_replace("\r","",$l["default"]);$yd=$l["generated"];return(in_array($yd,driver()->generated)?(JUSH=="mssql"?" AS ($j)".($yd=="VIRTUAL"?"":" $yd"):" GENERATED ALWAYS AS ($j) $yd"):(preg_match('~^GENERATED ~i',$j)?" $j":" DEFAULT ".(preg_match('~char|binary|text|json|enum|set|String~',$l["type"])||preg_match('~^(?![a-z])~i',$j)?(JUSH=="sql"&&preg_match('~text|json~',$l["type"])?"(".q($j).")":q($j)):str_ireplace("current_timestamp()","CURRENT_TIMESTAMP",(JUSH=="sqlite"?"($j)":$j)))));}function
type_class($U){foreach(array('char'=>'text','date'=>'time|year','binary'=>'blob','enum'=>'set',)as$x=>$X){if(preg_match("~$x|$X~",$U))return" class='$x'";}}function
edit_fields(array$m,array$gb,$U="TABLE",array$pd=array()){$m=array_values($m);$Qb=(($_POST?$_POST["defaults"]:get_setting("defaults"))?"":" class='hidden'");$kb=(($_POST?$_POST["comments"]:get_setting("comments"))?"":" class='hidden'");echo"<thead><tr>\n",($U=="PROCEDURE"?"<td>":""),"<th id='label-name'>".($U=="TABLE"?'Column name':'Parameter name'),"<td id='label-type'>".'Type'."<textarea id='enum-edit' rows='4' cols='12' wrap='off' hidden></textarea>".script("qs('#enum-edit').onblur = editingLengthBlur;"),"<td id='label-length'>".'Length',"<td>".'Options';if($U=="TABLE")echo"<td id='label-null'>NULL\n","<td><input type='radio' name='auto_increment_col' value=''><abbr id='label-ai' title='".'Auto Increment'."'>AI</abbr>",doc_link(array('sql'=>"example-auto-increment.html",'mariadb'=>"auto_increment/",)),"<td id='label-default'$Qb>".'Default value',(support("comment")?"<td id='label-comment'$kb>".'Comment':"");$Pe=!support("move_col");echo"<td>".icon("plus","add[".($Pe?count($m):0)."]","+",'Add next',($Pe?on('click','editingAddLastRow'):"")),"<tbody".on('click','editingClick').on('input','editingInput').on('keydown','editingKeydown').">\n";foreach($m
as$s=>$l){$s++;$yg=$l[($_POST?"orig":"field")];$cc=(isset($_POST["add"][$s-1])||(isset($l["field"])&&!idx($_POST["drop_col"],$s)))&&(support("drop_col")||$yg=="");echo"<tr".($cc?"":" hidden").">\n",($U=="PROCEDURE"?"<td>".html_select("fields[$s][inout]",explode("|",driver()->inout),$l["inout"]):"")."<th>",(support("move_col")?icon("move","","↕",'Move')." ":"");if($cc)echo"<input name='fields[$s][field]' value='".h($l["field"])."' data-maxlength='64' autocapitalize='off' aria-labelledby='label-name'".(isset($_POST["add"][$s-1])?" autofocus":"").">";echo
input_hidden("fields[$s][orig]",$yg);edit_type("fields[$s]",$l,$gb,$pd);if($U=="TABLE"){echo"<td><label class='block'>".checkbox("fields[$s][null]",1,$l["null"],"","","","label-null")."</label>","<td><label class='block'><input type='radio' name='auto_increment_col' value='$s'".($l["auto_increment"]?" checked":"")." aria-labelledby='label-ai'></label>","<td$Qb>".(driver()->generated?html_select("fields[$s][generated]",array_merge(array("","DEFAULT"),driver()->generated),$l["generated"])." ":checkbox("fields[$s][generated]",1,$l["generated"],"","","","label-default"));$b=" name='fields[$s][default]' aria-labelledby='label-default'";$Y=h($l["default"]);echo(preg_match('~\n~',$l["default"])?"<textarea$b rows='2' cols='30' style='vertical-align: bottom;'>\n$Y</textarea>":"<input$b value='$Y'>");if(support("comment")){$b=" name='fields[$s][comment]' data-maxlength='".(min_version(5.5)?1024:255)."' aria-labelledby='label-comment'";echo"<td$kb>".adminer()->commentInput('COLUMN',$b,$l["comment"]);}}echo"<td>",(support("move_col")?icon("plus","add[$s]","+",'Add next')." ":""),($yg==""||support("drop_col")?icon("cross","drop_col[$s]","x",'Remove'):"");}}function
process_fields(array&$m){if($_POST["add"]){$m=array_values($m);array_splice($m,key($_POST["add"]),0,array(array()));}return$_POST["add"]||$_POST["drop_col"];}function
normalize_enum(array$B){$X=$B[0];return"'".str_replace("'","''",addcslashes(stripcslashes(str_replace($X[0].$X[0],$X[0],substr($X,1,-1))),'\\'))."'";}function
grant($_d,array$qh,$d,$jg){if(!$qh)return
true;if($qh==array("ALL PRIVILEGES","GRANT OPTION"))return($_d=="GRANT"?queries("$_d ALL PRIVILEGES$jg WITH GRANT OPTION"):queries("$_d ALL PRIVILEGES$jg")&&queries("$_d GRANT OPTION$jg"));return
queries("$_d ".preg_replace('~(GRANT OPTION)\([^)]*\)~','\1',implode("$d, ",$qh).$d).$jg);}function
drop_create($nc,$g,$oc,$Wi,$pc,$A,$Af,$zf,$_f,$hg,$Rf){if($_POST["drop"])query_redirect($nc,$A,$Af);elseif($hg=="")query_redirect($g,$A,$_f);elseif(support("transaction_ddl")){driver()->begin();queries_redirect($A,$zf,queries($nc)&&queries($g)&&driver()->commit());driver()->rollback();}elseif($hg!=$Rf){$_b=queries($g);queries_redirect($A,$zf,$_b&&queries($nc));if($_b)queries($oc);}else
queries_redirect($A,$zf,queries($Wi)&&queries($pc)&&queries($nc)&&queries($g));}function
create_trigger($jg,array$K){$bj=" $K[Timing] $K[Event]".(preg_match('~ OF~',$K["Event"])?" $K[Of]":"");return"CREATE TRIGGER ".idf_escape($K["Trigger"]).(JUSH=="mssql"?$jg.$bj:$bj.$jg).rtrim(" $K[Type]\n$K[Statement]",";").";";}function
q_dollar($Q){$Ub='$$';while(strpos($Q.$Ub,$Ub)!=strlen($Q))$Ub='$_'.substr($Ub,1);return$Ub.$Q.$Ub;}function
create_routine($Ph,array$K){$O=array();$m=(array)$K["fields"];ksort($m);foreach($m
as$l){if($l["field"]!="")$O[]=(preg_match("~^(".driver()->inout.")\$~",$l["inout"])?"$l[inout] ":"").idf_escape($l["field"]).process_type($l,"CHARACTER SET");}$Sb=rtrim($K["definition"],";");return"CREATE $Ph ".idf_escape(trim($K["name"]))." (".implode(", ",$O).")".($Ph=="FUNCTION"?" RETURNS".process_type($K["returns"],"CHARACTER SET"):"").($K["language"]?" LANGUAGE $K[language]":"").(JUSH=="pgsql"?" AS ".q_dollar("\n".trim($Sb)."\n"):"\n$Sb;");}function
remove_definer($H){return
preg_replace('~^([A-Z =]+) DEFINER=`'.preg_replace('~@(.*)~','`@`(%|\1)',logged_user()).'`~','\1',$H);}function
format_foreign_key(array$o){$i=$o["db"];$Wf=$o["ns"];return" FOREIGN KEY (".implode(", ",array_map('Adminer\idf_escape',$o["source"])).") REFERENCES ".($i!=""&&$i!=$_GET["db"]?idf_escape($i).".":"").($Wf!=""&&$Wf!=$_GET["ns"]?idf_escape($Wf).".":"").idf_escape($o["table"])." (".implode(", ",array_map('Adminer\idf_escape',$o["target"])).")".(preg_match("~^(".driver()->onActions.")\$~",$o["on_delete"])?" ON DELETE $o[on_delete]":"").(preg_match("~^(".driver()->onActions.")\$~",$o["on_update"])?" ON UPDATE $o[on_update]":"").($o["deferrable"]?" $o[deferrable]":"");}function
tar_file($n,$gj){$J=pack("a100a8a8a8a12a12",$n,644,0,0,decoct($gj->size),decoct(time()));$Ya=8*32;for($s=0;$s<strlen($J);$s++)$Ya+=ord($J[$s]);$J
.=sprintf("%06o",$Ya)."\0 ";echo$J,str_repeat("\0",512-strlen($J));$gj->send();echo
str_repeat("\0",511-($gj->size+511)%512);}function
doc_link(array$Vg,$Xi="<sup>?</sup>"){$fi=connection()->server_info;$Wj=preg_replace('~^(\d\.?\d).*~s','\1',$fi);$Jj=array('sql'=>"https://dev.mysql.com/doc/refman/$Wj/en/",'sqlite'=>"https://www.sqlite.org/",'pgsql'=>"https://www.postgresql.org/docs/".(connection()->flavor=='cockroach'?"current":$Wj)."/",'mssql'=>"https://learn.microsoft.com/en-us/sql/",'oracle'=>"https://www.oracle.com/pls/topic/lookup?ctx=db".preg_replace('~^.* (\d+)\.(\d+)\.\d+\.\d+\.\d+.*~s','\1\2',$fi)."&id=",);if(connection()->flavor=='maria'){$Jj['sql']="https://mariadb.com/kb/en/";$Vg['sql']=(isset($Vg['mariadb'])?$Vg['mariadb']:str_replace(".html","/",$Vg['sql']));}return($Vg[JUSH]?"<a href='".h($Jj[JUSH].$Vg[JUSH].(JUSH=='mssql'?"?view=sql-server-ver$Wj":""))."'".target_blank().">$Xi</a>":"");}function
db_size($i){if(!connection()->select_db($i))return"?";$J=0;foreach(table_status()as$S)$J+=$S["Data_length"]+$S["Index_length"];return
format_number($J);}function
set_utf8mb4($g){static$O=false;if(!$O&&preg_match('~\butf8mb4~i',$g)){$O=true;echo"SET NAMES ".charset(connection()).";\n\n";}}if(isset($_GET["status"]))$_GET["variables"]=$_GET["status"];if(isset($_GET["import"]))$_GET["sql"]=$_GET["import"];if(DB==""&&isset($_GET["ns"]))redirect(remove_from_uri('ns'));if(!(DB!=""?connection()->select_db(DB):isset($_GET["sql"])||isset($_GET["dump"])||isset($_GET["database"])||isset($_GET["processlist"])||isset($_GET["privileges"])||isset($_GET["user"])||isset($_GET["variables"])||$_GET["script"]=="connect"||$_GET["script"]=="kill")){if(DB!=""||$_GET["refresh"]){restart_session();set_session("dbs",null);}if(DB!=""){header("HTTP/1.1 404 Not Found");page_header('Database'.": ".h(DB),'Invalid database.',true);}else{if($_POST["db"]&&!$k)queries_redirect(substr(ME,0,-1),'Databases have been dropped.',drop_databases($_POST["db"]));page_header('Select database',$k,false);echo"<p class='links'>\n";foreach(array('database'=>'Create database','privileges'=>'Privileges','processlist'=>'Process list','variables'=>'Variables','status'=>'Status',)as$x=>$X){if(support($x))echo"<a href='".h(ME)."$x='>$X</a>\n";}echo"<p>".sprintf('%s version: %s through PHP extension %s',get_driver(DRIVER),"<b>".h(connection()->server_info)."</b>","<b>".connection()->extension."</b>")."\n","<p>".sprintf('Logged as: %s',"<b>".h(logged_user())."</b>")."\n";$h=adminer()->databases();if($h){$Wh=support("scheme");$gb=collations();echo"<form action='' method='post'>\n","<table class='checkable odds'".on('click','tableClick').on('dblclick','tableClick').">\n","<thead><tr>".(support("database")?"<td class='hover'>":"")."<th".(JUSH!='mssql'?" aria-sort='ascending'":"").">".'Database'.(get_session("dbs")!==null?" - <a href='".h(ME)."refresh=1'>".'Refresh'."</a>":"")."<td>".'Collation'."<td>".'Tables'."<td>".'Size'." - <a href='".h(ME)."dbsize=1'".on('click','ajaxSetHtml',ME."script=connect").">".'Compute'."</a>"."<tbody>\n";$h=($_GET["dbsize"]?count_tables($h):array_flip($h));foreach($h
as$i=>$T){$Oh=h(ME)."db=".url_escape($i);$t=h("Db-".$i);echo"<tr>".(support("database")?"<td class='hover'>".checkbox("db[]",$i,in_array($i,(array)$_POST["db"]),"","","",$t):""),"<th><a href='$Oh' id='$t'>".h($i)."</a>";$fb=h(db_collation($i,$gb));echo"<td>".(support("database")?"<a href='$Oh".($Wh?"&amp;ns=":"")."&amp;database=' title='".'Alter database'."'>$fb</a>":$fb),"<td align='right'><a href='$Oh&amp;schema=' id='tables-".h($i)."' title='".'Database schema'."'>".($_GET["dbsize"]?$T:"?")."</a>","<td align='right' id='size-".h($i)."'>".($_GET["dbsize"]?db_size($i):"?"),"\n";}echo"</table>\n",(support("database")?"<div class='footer'><div>\n"."<fieldset><legend>".'Selected'." <span id='selected'></span></legend><div>\n"."<input type='hidden' name='all' value=''".on('click','countDbs').">\n"."<input type='submit' name='drop' value='".'Drop'."'".confirm().">\n"."</div></fieldset>\n"."</div></div>\n":""),input_token(),"</form>\n",script("tableCheck();");}$ha=adminer();$ch=($ha
instanceof
Plugins?$ha->plugins:array());$mc=($ha
instanceof
Plugins?$ha->drivers:array());$Zb=design_checksums();if($ch||$mc||$Zb){$Za=($ha
instanceof
Plugins?$ha->checksums():array());$ag=Plugins::officialChecksums();$Gj=function($Ij){return" (<a href='$Ij'".target_blank()." class='update'>".VERSION."</a>)";};$bh=function($cd)use($Za,$ag,$Gj){return($Za[$cd]&&$ag[$cd]&&$Za[$cd]!==$ag[$cd]?$Gj("https://www.adminer.org/plugins/?version=".VERSION):"");};echo"<div class='plugins'>\n","<h3>".'Loaded plugins'."</h3>\n<ul>\n";foreach($ch
as$ah){$Dh=new
\ReflectionObject($ah);$Wb=(method_exists($ah,'description')?$ah->description():"");if(!$Wb){if(preg_match('~^/[\s*]+(.+)~',$Dh->getDocComment(),$B))$Wb=$B[1];}$Xh=(method_exists($ah,'screenshot')?$ah->screenshot():"");echo"<li><b>".get_class($ah)."</b>".h($Wb?": $Wb":"").($Xh?" (<a href='".h($Xh)."'".target_blank().">".'screenshot'."</a>)":"").$bh(basename((string)$Dh->getFileName(),'.php'))."\n";}foreach($mc
as$t=>$D)echo"<li><b>".h($t)."</b>: ".h($D).$bh(basename((string)$ha->driverFiles[$t],'.php'))."\n";if($Zb){$cg=official_design_checksums();foreach($Zb
as$n=>$Yb){list($D,$Ya)=$Yb;$bg=$cg["$D/$n"];echo"<li><b>".h($n)."</b>".h($D?": $D":"").($bg&&$bg!==$Ya?$Gj("https://www.adminer.org/?version=".VERSION."#extras"):"")."\n";}}echo"</ul>\n";adminer()->pluginsLinks();echo"</div>\n";}}page_footer("db");exit;}adminer()->afterConnect();class
TmpFile{private$handler;var$size=0;function
__construct(){$this->handler=tmpfile();}function
write($sb){$this->size+=strlen($sb);fwrite($this->handler,$sb);}function
send(){fseek($this->handler,0);fpassthru($this->handler);fclose($this->handler);}}if(isset($_GET["select"])&&($_POST["edit"]||$_POST["clone"])&&!$_POST["save"])$_GET["edit"]=$_GET["select"];if(isset($_GET["callf"]))$_GET["call"]=$_GET["callf"];if(isset($_GET["function"]))$_GET["procedure"]=$_GET["function"];if(isset($_GET["download"])){$a=$_GET["download"];$m=fields($a);header("Content-Type: application/octet-stream");header("Content-Disposition: attachment; filename=".friendly_url("$a-".implode("_",$_GET["where"])).".".friendly_url($_GET["field"]));$M=array(idf_escape($_GET["field"]));$I=driver()->select($a,$M,array(where($_GET,$m)),$M);$K=($I?$I->fetch_row():array());echo
driver()->value($K[0],$m[$_GET["field"]]);exit;}elseif(isset($_GET["table"])){$a=$_GET["table"];$m=fields($a);if(!$m)$k=error()?:'No tables.';$S=table_status1($a);$D=adminer()->tableName($S);page_header(($m&&is_view($S)?$S['Engine']=='materialized view'?'Materialized view':'View':'Table').": ".($D!=""?$D:h($a)),$k);$Nh=array();foreach($m
as$x=>$l)$Nh+=$l["privileges"];adminer()->selectLinks($S,(isset($Nh["insert"])||!support("table")?"":null));$jb=$S["Comment"];if($jb!="")echo"<p class='nowrap'>".'Comment'.": ".adminer()->commentValue('TABLE',$jb)."\n";if($m)adminer()->tableStructurePrint($m,$S);function
tables_links(array$T){echo"<ul>\n";foreach($T
as$K){$_=preg_replace('~ns=[^&]*~',"ns=".url_escape($K["ns"]),ME);echo"<li><a href='".h($_."table=".url_escape($K["table"]))."'>".($K["ns"]!=$_GET["ns"]?"<b>".h($K["ns"])."</b>.":"").h($K["table"])."</a>";}echo"</ul>\n";}$ne=driver()->inheritsFrom($a);if($ne){echo"<h3>".'Inherits from'."</h3>\n";tables_links($ne);}if(support("indexes")&&driver()->supportsIndex($S)){echo"<div>\n","<h3 id='indexes'>".'Indexes'."</h3>\n";$w=indexes($a);if($w)adminer()->tableIndexesPrint($w,$S);if(driver()->supportsAlterIndex($S))echo'<p class="links hover"><a href="'.h(ME).'indexes='.url_escape($a).'">'.'Alter indexes'."</a>\n";echo"</div>\n";}if(!is_view($S)){if(fk_support($S)){echo"<div>\n","<h3 id='foreign-keys'>".'Foreign keys'."</h3>\n";$pd=foreign_keys($a);if($pd){echo"<table>\n","<thead><tr><th>".'Source'."<td>".'Target'."<td>".'ON DELETE'."<td>".'ON UPDATE'."<td class='hover'><tbody>\n";foreach($pd
as$D=>$o){echo"<tr title='".h($D)."'>","<th><i>".implode("</i>, <i>",array_map('Adminer\h',$o["source"]))."</i>";$_=($o["db"]!=""?preg_replace('~db=[^&]*~',"db=".url_escape($o["db"]),ME):($o["ns"]!=""?preg_replace('~ns=[^&]*~',"ns=".url_escape($o["ns"]),ME):ME));echo"<td><a href='".h($_."table=".url_escape($o["table"]))."'>".($o["db"]!=""&&$o["db"]!=DB?"<b>".h($o["db"])."</b>.":"").($o["ns"]!=""&&$o["ns"]!=$_GET["ns"]?"<b>".h($o["ns"])."</b>.":"").h($o["table"])."</a>","(<i>".implode("</i>, <i>",array_map('Adminer\h',$o["target"]))."</i>)","<td>".h($o["on_delete"]),"<td>".h($o["on_update"]),'<td class="hover"><a href="'.h(ME.'foreign='.url_escape($a).'&name='.url_escape($D)).'">'.'Alter'.'</a>',"\n";}echo"</table>\n";}echo'<p class="links hover"><a href="'.h(ME).'foreign='.url_escape($a).'">'.'Create foreign key'."</a>\n","</div>\n";}if(support("check")){echo"<div>\n","<h3 id='checks'>".'Checks'."</h3>\n";$Va=driver()->checkConstraints($a);if($Va){echo"<table>\n";foreach($Va
as$x=>$X)echo"<tr title='".h($x)."'>","<td><code class='jush-".JUSH."'>".shorten_utf8(preg_replace('~\s+~',' ',ltrim($X)),80,"</code>"),"<td class='hover'><a href='".h(ME.'check='.url_escape($a).'&name='.url_escape($x))."'>".'Alter'."</a>","\n";echo"</table>\n";}echo'<p class="links hover"><a href="'.h(ME).'check='.url_escape($a).'">'.'Create check'."</a>\n","</div>\n";}}if(support(is_view($S)?"view_trigger":"trigger")){echo"<div>\n","<h3 id='triggers'>".'Triggers'."</h3>\n";$sj=triggers($a);if($sj){echo"<table>\n";foreach($sj
as$x=>$X)echo"<tr valign='top'><td>".h($X[0])."<td>".h($X[1])."<th>".h($x)."<td class='hover'><a href='".h(ME.'trigger='.url_escape($a).'&name='.url_escape($x))."'>".'Alter'."</a>\n";echo"</table>\n";}echo'<p class="links hover"><a href="'.h(ME).'trigger='.url_escape($a).'">'.'Create trigger'."</a>\n","</div>\n";}$me=driver()->inheritedTables($a);if($me){echo"<h3 id='partitions'>".'Inherited by'."</h3>\n";$Mg=driver()->partitionsInfo($a);if($Mg)echo"<p><code class='jush-".JUSH."'>BY ".h("$Mg[partition_by]($Mg[partition])")."</code>\n";tables_links($me);}}elseif(isset($_GET["schema"])){page_header('Database schema',"",array(),h(DB.($_GET["ns"]?".$_GET[ns]":"")));$Mi=array();$Ni=array();$Zc=array();$ca=($_GET["schema"]?:$_COOKIE["adminer_schema-".str_replace(".","_",DB)]);preg_match_all('~([^:]+):([-0-9.]+)x([-0-9.]+)(_|$)~',$ca,$if,PREG_SET_ORDER);foreach($if
as$s=>$B){$Mi[$B[1]]=array((float)$B[2],(float)$B[3]);$Ni[]="\n\t'".js_escape($B[1])."': [ $B[2], $B[3] ]";}$jj=0;$Ga=-1;$Vh=array();$Ch=array();$Te=array();$na=driver()->allFields();foreach(table_status('',true)as$R=>$S){if(is_view($S))continue;$G=0;$Vh[$R]["fields"]=array();foreach($na[$R]as$l){$G+=1.25;$Zc[$R][$l["field"]]=$G;$Vh[$R]["fields"][$l["field"]]=$l;}$Vh[$R]["pos"]=($Mi[$R]?:array($jj,0));foreach(adminer()->foreignKeys($R)as$X){if(!$X["db"]){$Re=$Ga;if(idx($Mi[$R],1)||idx($Mi[$X["table"]],1))$Re=min(idx($Mi[$R],1,0),idx($Mi[$X["table"]],1,0))-1;else$Ga-=.1;while($Te[(string)$Re])$Re-=.0001;$Vh[$R]["references"][$X["table"]][(string)$Re]=array($X["source"],$X["target"]);$Ch[$X["table"]][$R][(string)$Re]=$X["target"];$Te[(string)$Re]=true;}}$jj=max($jj,$Vh[$R]["pos"][0]+2.5+$G);}echo'<div id="schema" style="height: ',$jj,'em;">
<script',nonce(),'>
const tablePos = {',implode(",",$Ni)."\n",'};
const em = qs(\'#schema\').offsetHeight / ',$jj,';
document.onmousemove = schemaMousemove;
document.onmouseup = event => schemaMouseup(event, \'',js_escape(DB),'\');
</script>
';foreach($Vh
as$D=>$R){echo"<div class='table'".on('mousedown','schemaMousedown')." style='top: ".$R["pos"][0]."em; left: ".$R["pos"][1]."em;'>",'<a href="'.h(ME).'table='.url_escape($D).'"><b>'.h($D)."</b></a>";foreach($R["fields"]as$l){$X='<span'.type_class($l["type"]).' title="'.h($l["type"].($l["length"]?"($l[length])":"").($l["null"]?" NULL":'')).'">'.h($l["field"]).'</span>';echo"<br>".($l["primary"]?"<i>$X</i>":$X);}foreach((array)$R["references"]as$Ui=>$Eh){foreach($Eh
as$Re=>$_h){$Se=$Re-idx($Mi[$D],1);$s=0;foreach($_h[0]as$ri)echo"\n<div class='references' title='".h($Ui)."' id='refs$Re-".($s++)."' style='left: $Se"."em; top: ".$Zc[$D][$ri]."em; padding-top: .5em;'>"."<div style='border-top: 1px solid gray; width: ".(-$Se)."em;'></div></div>";}}foreach((array)$Ch[$D]as$Ui=>$Eh){foreach($Eh
as$Re=>$d){$Se=$Re-idx($Mi[$D],1);$s=0;foreach($d
as$Ti)echo"\n<div class='references arrow' title='".h($Ui)."' id='refd$Re-".($s++)."' style='left: $Se"."em; top: ".$Zc[$D][$Ti]."em;'>"."<div style='height: .5em; border-bottom: 1px solid gray; width: ".(-$Se)."em;'></div>"."</div>";}}echo"\n</div>\n";}foreach($Vh
as$D=>$R){foreach((array)$R["references"]as$Ui=>$Eh){if($Vh[$Ui]){foreach($Eh
as$Re=>$_h){$Gf=$jj;$qf=-10;foreach($_h[0]as$x=>$ri){$eh=$R["pos"][0]+$Zc[$D][$ri];$fh=$Vh[$Ui]["pos"][0]+$Zc[$Ui][$_h[1][$x]];$Gf=min($Gf,$eh,$fh);$qf=max($qf,$eh,$fh);}echo"<div class='references' id='refl$Re' style='left: $Re"."em; top: $Gf"."em; padding: .5em 0;'><div style='border-right: 1px solid gray; margin-top: 1px; height: ".($qf-$Gf)."em;'></div></div>\n";}}}}echo'</div>
<p class="links"><a href="',h(ME."schema=".url_escape($ca)),'" id="schema-link">Permanent link</a>
';}elseif(isset($_GET["dump"])){$a=$_GET["dump"];if($_POST&&!$k){$j=array("auto_increment"=>'');foreach(array("type","routine","event","trigger")as$Hi){if(support($Hi))$j[$Hi."s"]='';}save_settings(array_intersect_key($_POST+$j,array_flip(array("output","format","db_style","table_style","data_style"))+$j),"adminer_export");$T=array_flip((array)$_POST["tables"])+array_flip((array)$_POST["data"]);$Qc=dump_headers((count($T)==1?key($T):DB),(DB==""||$_GET["ns"]===""||count($T)>1));$Ae=preg_match('~sql~',$_POST["format"]);if($Ae){echo"-- Adminer ".VERSION." ".get_driver(DRIVER)." ".str_replace("\n"," ",connection()->server_info)." dump\n\n";if(JUSH=="sql"){echo"SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
".($_POST["data_style"]?"SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
":"")."
";connection()->query("SET time_zone = '+00:00'");connection()->query("SET sql_mode = ''");}}$Ci=$_POST["db_style"];$h=array(DB);if(DB==""){$h=$_POST["databases"];if(is_string($h))$h=explode("\n",rtrim(str_replace("\r","",$h),"\n"));}foreach((array)$h
as$i){adminer()->dumpDatabase($i);if(connection()->select_db($i)){if($Ae&&$Ci)echo
use_sql($i,$Ci).";\n\n";foreach(($_GET["ns"]===""?(array)$_POST["schemas"]:(DB!=""||!support("scheme")?array(""):adminer()->schemas()))as$Vh){if($Vh!=""){if(DB==""&&information_schema(DB,$Vh))continue;set_schema($Vh);}$_i=($_POST["table_style"]||$_POST["data_style"]?table_status('',true):array());$Pc=array();$Jb=array();foreach($_i
as$D=>$S){if(DB==""||$_GET["ns"]===""||in_array($D,(array)$_POST["tables"]))$Pc[$D]=$S;if(DB==""||$_GET["ns"]===""||in_array($D,(array)$_POST["data"]))$Jb[$D]=$S;}if($Ae){if($_POST["table_style"]=="DROP+CREATE"&&function_exists('Adminer\drop_sql'))echo
drop_sql($Pc);if($_POST["data_style"]=="TRUNCATE+INSERT"&&function_exists('Adminer\truncate_all_sql')){$tj=array();foreach($Jb
as$D=>$S){if(!is_view($S)&&!($_POST["table_style"]=="DROP+CREATE"&&isset($Pc[$D])))$tj[]=$D;}echo
truncate_all_sql($tj);}$Eg="";if($_POST["types"]){foreach(types()as$t=>$U){$Sb=type_definition($t);$Zf=($Sb["kind"]=='d'?"DOMAIN":"TYPE");if($Sb["definition"])$Eg
.=($Ci!='DROP+CREATE'?"DROP $Zf IF EXISTS ".idf_escape($U).";;\n":"")."CREATE $Zf ".idf_escape($U)." $Sb[definition];\n\n";else$Eg
.="-- Could not export type $U\n\n";}}if($_POST["routines"]){foreach(routines()as$K){$D=$K["ROUTINE_NAME"];$Ph=$K["ROUTINE_TYPE"];$g=create_routine($Ph,array("name"=>$D)+routine($K["SPECIFIC_NAME"],$Ph));set_utf8mb4($g);$Eg
.=($Ci!='DROP+CREATE'?"DROP $Ph IF EXISTS ".idf_escape($D).";;\n":"")."$g;\n\n";}}if($_POST["events"]){foreach(get_rows("SHOW EVENTS",null,"-- ")as$K){$g=remove_definer(get_val("SHOW CREATE EVENT ".idf_escape($K["Name"]),3));set_utf8mb4($g);$Eg
.=($Ci!='DROP+CREATE'?"DROP EVENT IF EXISTS ".idf_escape($K["Name"]).";;\n":"")."$g;;\n\n";}}echo($Eg&&JUSH=='sql'?"DELIMITER ;;\n\n$Eg"."DELIMITER ;\n\n":$Eg);}if($_POST["table_style"]||$_POST["data_style"]){$Yj=array();foreach($_i
as$D=>$S){$R=array_key_exists($D,$Pc);$Hb=array_key_exists($D,$Jb);if($R||$Hb){$gj=null;if($Qc=="tar"){$gj=new
TmpFile;ob_start(array($gj,'write'),1e5);}adminer()->dumpTable($D,($R?$_POST["table_style"]:""),(is_view($S)?2:0));if(is_view($S))$Yj[]=$D;elseif($Hb){$m=fields($D);$M=array("*");$vb=convert_fields($m,$m);if($vb)$M[]=substr($vb,2);adminer()->dumpData($D,$_POST["data_style"],"",$M);}if($Ae&&$_POST["triggers"]&&$R&&($sj=trigger_sql($D)))echo"\nDELIMITER ;;\n$sj\nDELIMITER ;\n";if($Qc=="tar"){ob_end_flush();tar_file((DB!=""?"":"$i/")."$D.csv",$gj);}elseif($Ae)echo"\n";}}if($Ae&&$_POST["table_style"]&&function_exists('Adminer\foreign_keys_sql')){foreach($Pc
as$D=>$S){if(!is_view($S))echo
foreign_keys_sql($D);}}if($Ae){foreach($Yj
as$Xj)adminer()->dumpTable($Xj,$_POST["table_style"],1);}if($Qc=="tar")echo
pack("x1024");}}}}adminer()->dumpFooter();exit;}page_header('Export',$k,($_GET["export"]!=""?array("table"=>$_GET["export"]):array()),h(DB));echo'
<form action="" method="post">
<table class="layout">
';$Mb=array('','USE','DROP+CREATE','CREATE');$Oi=array('','DROP+CREATE','CREATE');$Ib=array('','TRUNCATE+INSERT','INSERT');if(JUSH=="sql")$Ib[]='INSERT+UPDATE';$K=get_settings("adminer_export");if(!$K)$K=array("output"=>"text","format"=>"sql","db_style"=>(DB!=""?"":"CREATE"),"table_style"=>"DROP+CREATE","data_style"=>"INSERT");echo"<tr><th>".'Output'."<td>".html_radios("output",adminer()->dumpOutput(),$K["output"])."\n","<tr><th>".'Format'."<td>".html_radios("format",adminer()->dumpFormat(),$K["format"])."\n",(JUSH=="sqlite"?"":"<tr><th>".'Database'."<td>".html_select('db_style',$Mb,$K["db_style"]).(support("type")?checkbox("types",1,$K["types"],'User types'):"").(support("routine")?checkbox("routines",1,$K["routines"],'Routines'):"").(support("event")?checkbox("events",1,$K["events"],'Events'):"")),"<tr><th>".'Tables'."<td>".html_select('table_style',$Oi,$K["table_style"]).checkbox("auto_increment",1,$K["auto_increment"],'Auto Increment').(support("trigger")?checkbox("triggers",1,$K["triggers"],'Triggers'):""),"<tr><th>".'Data'."<td>".html_select('data_style',$Ib,$K["data_style"]),'</table>
<p><input type=\'submit\' value=\'Export\'>
',input_token(),'
<table',on('click','dumpClick'),'>
';$lh=array();if($_GET["ns"]===""){echo"<thead><tr><th style='text-align: left;'>","<label class='block'><input type='checkbox' id='check-schemas' checked class='jsonly' title='".'All'."'".on('click','formCheck','^schemas\[').">".'Schema'."</label>","<tbody>\n";foreach(adminer()->schemas()as$Vh){if(!information_schema(DB,$Vh))echo"<tr><td>".checkbox("schemas[]",$Vh,true,$Vh,"","block")."\n";}}elseif(DB!=""){$Wa=($a!=""?"":" checked");echo"<thead><tr>","<th style='text-align: left;'><label class='block'><input type='checkbox' id='check-tables'$Wa class='jsonly' title='".'All'."'".on('click','formCheck','^tables\[').">".'Table'."</label>","<th style='text-align: right;'><label class='block'>".'Data'."<input type='checkbox' id='check-data'$Wa class='jsonly' title='".'All'."'".on('click','formCheck','^data\[')."></label>","<tbody>\n";$Yj="";$Qi=tables_list();foreach($Qi
as$D=>$U){$kh=preg_replace('~_.*~','',$D);$Wa=($a==""||$a==(substr($a,-1)=="%"?"$kh%":$D));$nh="<tr><td>".checkbox("tables[]",$D,$Wa,$D,"","block");if($U!==null&&!preg_match('~table~i',$U))$Yj
.="$nh\n";else
echo"$nh<td align='right'><label class='block'><span id='Rows-".h($D)."'></span>".checkbox("data[]",$D,$Wa)."</label>\n";$lh[$kh]++;}echo$Yj;if($Qi)echo
script("ajaxSetHtml('".js_escape(ME)."script=db');");}else{$h=adminer()->databases();echo"<thead><tr><th style='text-align: left;'>","<label class='block'>".($h?"<input type='checkbox' id='check-databases'".($a==""?" checked":"")." class='jsonly' title='".'All'."'".on('click','formCheck','^databases\[').">":"").'Database'."</label>","<tbody>\n";if($h){foreach($h
as$i){if(!information_schema($i)){$kh=preg_replace('~_.*~','',$i);echo"<tr><td>".checkbox("databases[]",$i,$a==""||$a=="$kh%",$i,"","block")."\n";$lh[$kh]++;}}}else
echo"<tr><td><textarea name='databases' rows='10' cols='20'></textarea>";}echo'</table>
</form>
';$id=true;foreach($lh
as$x=>$X){if($x!=""&&$X>1){echo($id?"<p>":" ")."<a href='".h(ME)."dump=".url_escape("$x%")."'>".h($x)."</a>";$id=false;}}}elseif(isset($_GET["privileges"])){page_header('Privileges');echo'<p class="links"><a href="'.h(ME).'user=">'.'Create user'."</a>";$I=connection()->query("SELECT User, Host FROM mysql.".(DB==""?"user":"db WHERE ".q(DB)." LIKE Db")." ORDER BY Host, User");$_d=$I;if(!$I)$I=connection()->query("SELECT SUBSTRING_INDEX(CURRENT_USER, '@', 1) AS User, SUBSTRING_INDEX(CURRENT_USER, '@', -1) AS Host");echo"<form action=''><p>\n";hidden_fields_get();echo
input_hidden("db",DB),($_d?"":input_hidden("grant")),"<table class='odds'>\n","<thead><tr><th>".'Username'."<th>".'Server'."<td class='hover'><tbody>\n";while($K=$I->fetch_assoc())echo'<tr><td>'.h($K["User"]),"<td>".h($K["Host"]),'<td class="hover"><a href="'.h(ME.'user='.url_escape($K["User"]).'&host='.url_escape($K["Host"])).'">'.'Edit'."</a>\n";if(!$_d||DB!="")echo"<tr><td><input name='user' autocapitalize='off'><td><input name='host' value='localhost' autocapitalize='off'><td><input type='submit' value='".'Edit'."'>\n";echo"</table>\n","</form>\n";}elseif(isset($_GET["sql"])){if(!$k&&$_POST["export"]){save_settings(array("output"=>$_POST["output"],"format"=>$_POST["format"]),"adminer_import");dump_headers("sql");if($_POST["format"]=="sql")echo"$_POST[query]\n";else{adminer()->dumpTable("","");adminer()->dumpData("","table",$_POST["query"]);adminer()->dumpFooter();}exit;}restart_session();$Sd=&get_session("queries");$Rd=&$Sd[DB];if(!$k&&$_POST["clear"]){$Rd=array();redirect(remove_from_uri("history"));}stop_session();$ia=get_settings("adminer_import");if($_POST&&$ia)save_settings($ia,"adminer_import");page_header((isset($_GET["import"])?'Import':'SQL command'),$k);$Ze=driver()->lineComment();if(!$k&&$_POST&&!(isset($_GET["import"])&&adminer()->importProcess())){$Ub=driver()->delimiter;$p=false;if(!isset($_GET["import"]))$H=$_POST["query"];elseif($_POST["webfile"]){$ui=adminer()->importServerPath();$p=@fopen((file_exists($ui)?$ui:"compress.zlib://$ui.gz"),"rb");$H=($p?fread($p,1e6):false);}else$H=get_file("sql_file",true,$Ub);if(is_string($H)){if(($xf=ini_bytes("memory_limit"))!="-1")ini_set("memory_limit",max($xf,strval(2*strlen($H)+memory_get_usage()+8e6)));if($H!=""&&strlen($H)<1e6){$th=$H.(preg_match("~$Ub\\s*\$~",$H)?"":$Ub);if(!$Rd||first(end($Rd))!=$th){restart_session();$Rd[]=array($th,time());set_session("queries",$Sd);stop_session();}}$si="(?:\\s|/\\*[\s\S]*?\\*/|(?:$Ze)[^\n]*\n?|--\r?\n)";$dg=0;$yc=true;$xb=false;$f=connect();if($f&&DB!=""){$f->select_db(DB);if($_GET["ns"]!="")set_schema($_GET["ns"],$f);}$ib=0;$Ec=array();$Kg='[\'"'.(JUSH=="sql"?'`':(JUSH=="sqlite"?'`[':(JUSH=="mssql"?'[':''))).']|/\*|'.$Ze.'|$'.(JUSH=="pgsql"?'|\$([a-zA-Z]\w*)?\$':'');$kj=microtime(true);while($H!=""){if(!$dg&&preg_match("~^$si*+DELIMITER\\s+(\\S+)~i",$H,$B)){$Ub=preg_quote($B[1]);$H=substr($H,strlen($B[0]));}elseif(!$dg&&JUSH=='pgsql'&&preg_match("~^($si*+COPY\\s+)[^;]+\\s+FROM\\s+stdin;~i",$H,$B)){$Ub="\n\\\\\\.\r?\n";$xb=true;$dg=strlen($B[0]);}else{preg_match("($Ub\\s*|$Kg)",$H,$B,PREG_OFFSET_CAPTURE,$dg);list($rd,$G)=$B[0];if(!$rd&&$p&&!feof($p))$H
.=fread($p,1e5);else{if(!$rd&&rtrim($H)=="")break;$dg=$G+strlen($rd);if($rd&&!preg_match("(^$Ub)",$rd)){$Pa=driver()->hasCStyleEscapes()||(JUSH=="pgsql"&&($G>0&&strtolower($H[$G-1])=="e"));$Wg=($rd=='/*'?'\*/':($rd=='['?']':(preg_match("~^(?:$Ze)~",$rd)?"\n":preg_quote($rd).($Pa?'|\\\\.':''))));while(preg_match("($Wg|\$)s",$H,$B,PREG_OFFSET_CAPTURE,$dg)){$Th=$B[0][0];if(!$Th&&$p&&!feof($p))$H
.=fread($p,1e5);else{$dg=$B[0][1]+strlen($Th);if(!$Th||$Th[0]!="\\")break;}}}else{$yc=false;$th=substr($H,0,$G+($xb?3:0));$ib++;$nh="<pre id='sql-$ib'><code class='jush-".JUSH."'>".adminer()->sqlCommandQuery($th)."</code></pre>\n";if(JUSH=="sqlite"&&preg_match("~^$si*+(ATTACH|VACUUM\\b.*\\bINTO)\\b~is",$th,$B)!==0){echo$nh,"<p class='error'>".sprintf('%s queries are not supported.',preg_match('~ATTACH~i',$B[1])?'ATTACH':'VACUUM INTO')."\n";$Ec[]=" <a href='#sql-$ib'>$ib</a>";if($_POST["error_stops"])break;}else{if(!$_POST["only_errors"]){echo$nh;ob_flush();flush();}$zi=microtime(true);if(connection()->multi_query($th)&&$f&&preg_match("~^$si*+USE\\b~i",$th))$f->query($th);do{$I=connection()->store_result();if(connection()->error){echo($_POST["only_errors"]?$nh:""),"<p class='error'>".'Error in query'.(connection()->errno?" (".connection()->errno.")":"").": ".error()."\n";$Ec[]=" <a href='#sql-$ib'>$ib</a>";if($_POST["error_stops"])break
2;}else{$_=ME."sql=".url_escape(trim($th));$Zi=" <span class='time'>(".format_time($zi).")</span>".(strlen($_)<1900?" <a href='".h($_)."'>".'Edit'."</a>":"");$ka=connection()->affected_rows;$bk=($_POST["only_errors"]?"":driver()->warnings());$ck="warnings-$ib";if($bk)$Zi
.=", <a href='#$ck' class='toggle'>".'Warnings'."</a>";$Nc=null;$xg=null;$Oc="explain-$ib";if(is_object($I)){$z=$_POST["limit"];$Xf=$z;$xg=print_select_result($I,$f,array(),$Xf);if(!$_POST["only_errors"]){echo"<form action='' method='post'>\n";$Xf=max($I->num_rows,$Xf);echo"<p class='sql-footer'>".($Xf?($z&&$Xf>$z?sprintf('%d / ',$z):"").lang_format(array('%d row','%d rows'),$Xf):""),$Zi;if($f&&preg_match("~^($si|\\()*+SELECT\\b~i",$th)&&($Nc=explain($f,$th)))echo", <a href='#$Oc' class='toggle'>Explain</a>";$t="export-$ib";echo", <a href='#$t' class='toggle'>".'Export'."</a><span id='$t' class='hidden'>: ".html_select("output",adminer()->dumpOutput(),$ia["output"])." ".html_select("format",adminer()->dumpFormat(),$ia["format"]).input_hidden("query",$th)."<input type='submit' name='export' value='".'Export'."'".($z?"":on('click','sqlExport')).">".input_token()."</span>\n"."</form>\n";}}else{if(preg_match("~^$si*+(CREATE|DROP|ALTER)$si++(DATABASE|SCHEMA)\\b~i",$th)){restart_session();set_session("dbs",null);stop_session();}if(!$_POST["only_errors"])echo"<p class='message' title='".h(connection()->info)."'>".lang_format(array('Query executed OK, %d row affected.','Query executed OK, %d rows affected.'),$ka)."$Zi\n";}echo($bk?"<div id='$ck' class='hidden'>\n$bk</div>\n":"");if($Nc){echo"<div id='$Oc' class='hidden explain'>\n";print_select_result($Nc,$f,$xg);echo"</div>\n";}}$zi=microtime(true);}while(connection()->next_result());}$H=substr($H,$dg);$dg=0;if($xb){$Ub=driver()->delimiter;$xb=false;}}}}}if($yc)echo"<p class='message'>".'No commands to execute.'."\n";else{$fe=connection()->inTransaction();driver()->rollback();if($fe)echo"<pre><code class='jush-".JUSH."'>ROLLBACK -- Adminer</code></pre>\n";if($_POST["only_errors"])echo"<p class='message'>".lang_format(array('%d query executed OK.','%d queries executed OK.'),$ib-count($Ec))," <span class='time'>(".format_time($kj).")</span>\n";elseif($Ec&&$ib>1)echo"<p class='error'>".'Error in query'.": ".implode("",$Ec)."\n";}}else
echo"<p class='error'>".upload_error($H)."\n";}echo'
<form action="" method="post" enctype="multipart/form-data" id="form"',(isset($_GET["import"])?"":on('submit','sqlSubmit',remove_from_uri("sql|limit|error_stops|only_errors|history"))),'>
';$Lc="<input type='submit' value='".'Execute'."' title='Ctrl+Enter'>";if(!isset($_GET["import"])){$th=$_GET["sql"];if($_POST)$th=$_POST["query"];elseif($_GET["history"]=="all")$th=$Rd;elseif($_GET["history"]!="")$th=idx($Rd[$_GET["history"]],0);echo"<p>";textarea("query",$th,20);echo($_POST?"":script("qs('textarea').focus();")),"<p>";adminer()->sqlPrintAfter();echo"$Lc\n",'Limit rows'.": <input type='number' name='limit' class='size' value='".h($_POST?$_POST["limit"]:$_GET["limit"])."'>\n";}else{$Ed=(extension_loaded("zlib")?"[.gz]":"");echo"<fieldset><legend>".'File upload'."</legend><div>","SQL$Ed: ".file_input(" name='sql_file[]' multiple","\n$Lc"),"</div></fieldset>\n";$ce=adminer()->importServerPath();if($ce)echo"<fieldset><legend>".'From server'."</legend><div>",sprintf('Webserver file %s',"<code>".h($ce)."$Ed</code>")," <input type='submit' name='webfile' value='".'Run file'."'>","</div></fieldset>\n";adminer()->importPrint();echo"<p>";}echo
checkbox("error_stops",1,($_POST?$_POST["error_stops"]:isset($_GET["import"])||$_GET["error_stops"]),'Stop on error')."\n",checkbox("only_errors",1,($_POST?$_POST["only_errors"]:isset($_GET["import"])||$_GET["only_errors"]),'Show only errors')."\n",input_token();if(!isset($_GET["import"])&&$Rd){print_fieldset("history",'History',$_GET["history"]!="");for($X=end($Rd);$X;$X=prev($Rd)){$x=key($Rd);list($th,$Zi,$uc)=$X;echo'<div><a href="'.h(ME."sql=&history=$x").'" class="hover">'.'Edit'."</a>"." <span class='time' title='".@date('Y-m-d',$Zi)."'>".@date("H:i:s",$Zi)."</span>"." <code class='jush-".JUSH."'>".shorten_utf8(preg_replace('~\s+~',' ',ltrim(preg_replace("~^(?:$Ze).*~m",'',$th))),80,"</code>").($uc?" <span class='time'>($uc)</span>":"")."</div>\n";}echo"<input type='submit' name='clear' value='".'Clear'."'>\n","<a href='".h(ME."sql=&history=all")."'>".'Edit all'."</a>\n","</div></fieldset>\n";}echo'</form>
';}elseif(isset($_GET["edit"])){$a=$_GET["edit"];$m=fields($a);$Z=(isset($_GET["select"])?($_POST["check"]&&count($_POST["check"])==1?where_check($_POST["check"][0],$m):""):where($_GET,$m));$Fj=(isset($_GET["select"])?$_POST["edit"]:$Z);foreach($m
as$D=>$l){if((!$Fj&&!isset($l["privileges"]["insert"]))||adminer()->fieldName($l)=="")unset($m[$D]);}if($_POST&&!$k&&!isset($_GET["select"])){$A=$_POST["referer"];if($_POST["insert"])$A=($Fj?null:$_SERVER["REQUEST_URI"]);elseif(!preg_match('~^.+&select=.+$~',$A))$A=ME."select=".url_escape($a);$w=indexes($a);$_j=unique_array($_GET["where"],$w);$wh="\nWHERE $Z";if(isset($_POST["delete"]))queries_redirect($A,'Item has been deleted.',driver()->delete($a,$wh,$_j?0:1));else{$O=array();foreach($m
as$D=>$l){$X=process_input($l);if($X!==false&&$X!==null)$O[idf_escape($D)]=$X;}if($Fj){if(!$O)redirect($A);queries_redirect($A,'Item has been updated.',driver()->update($a,$O,$wh,$_j?0:1));if(is_ajax()){page_headers();page_messages($k);exit;}}else{$I=driver()->insert($a,$O);$Qe=($I?last_id($I):0);queries_redirect($A,sprintf('Item%s has been inserted.',($Qe?" $Qe":"")),$I);}}}$K=null;if($Z){$M=array();foreach($m
as$D=>$l){if(isset($l["privileges"]["select"])){$va=($_POST["clone"]&&$l["auto_increment"]?"''":convert_field($l));$M[]=($va?"$va AS ":"").idf_escape($D);}}$K=array();if(!support("table"))$M=array("*");if($M){$I=driver()->select($a,$M,array($Z),$M,array(),(isset($_GET["select"])?2:1));if(!$I)$k=error();else{$K=$I->fetch_assoc();if(!$K)$K=false;}if(isset($_GET["select"])&&(!$K||$I->fetch_assoc()))$K=null;}}if(!$m&&driver()->primary!=""){if(!$Z){$I=driver()->select($a,array("*"),array(),array("*"));$K=($I?$I->fetch_assoc():false);if(!$K)$K=array(driver()->primary=>"");}if($K){foreach($K
as$x=>$X){if(!$Z)$K[$x]=null;$m[$x]=array("field"=>$x,"null"=>($x!=driver()->primary),"auto_increment"=>($x==driver()->primary));}}}if($_POST["save"]){$gh=array();foreach((array)$_POST["fields"]as$x=>$X)$gh[bracket_escape($x,true)]=$X;$K=$gh+($K?$K:array());}edit_form($a,$m,$K,$Fj,$k);}elseif(isset($_GET["create"])){$a=$_GET["create"];$Og=driver()->partitionBy;$Sg=($Og&&$a!=""?driver()->partitionsInfo($a):array());$Bh=referencable_primary($a);$pd=array();foreach($Bh
as$Li=>$l)$pd[str_replace("`","``",$Li)."`".str_replace("`","``",$l["field"])]=$Li;$_g=array();$S=array();if($a!=""){$_g=fields($a);$S=table_status1($a);if(count($S)<2)$k='No tables.';}$K=$_POST;$K["fields"]=(array)$K["fields"];if($K["auto_increment_col"])$K["fields"][$K["auto_increment_col"]]["auto_increment"]=true;if($_POST&&!$k)save_settings(array("comments"=>$_POST["comments"],"defaults"=>$_POST["defaults"]));if($_POST&&!process_fields($K["fields"])&&!$k){if($_POST["drop"])queries_redirect(substr(ME,0,-1),'Table has been dropped.',drop_tables(array($a)));else{$m=array();$na=array();$Kj=false;$nd=array();$zg=reset($_g);$ma=" FIRST";foreach($K["fields"]as$x=>$l){$o=$pd[$l["type"]];$vj=($o!==null?$Bh[$o]:$l);if($l["field"]!=""){if(!$l["generated"])$l["default"]=null;$sh=process_field($l,$vj);$na[]=array($l["orig"],$sh,$ma);if(!$zg||$sh!==process_field($zg,$zg)){$m[]=array($l["orig"],$sh,$ma);if($l["orig"]!=""||$ma)$Kj=true;}if($o!==null)$nd[idf_escape($l["field"])]=($a!=""&&JUSH!="sqlite"?"ADD":" ").format_foreign_key(array('table'=>$pd[$l["type"]],'source'=>array($l["field"]),'target'=>array($vj["field"]),'on_delete'=>$l["on_delete"],));$ma=" AFTER ".idf_escape($l["field"]);}elseif($l["orig"]!=""){$Kj=true;$m[]=array($l["orig"]);}if($l["orig"]!=""){$zg=next($_g);if(!$zg)$ma="";}}$Qg=array();if(in_array($K["partition_by"],$Og)){foreach($K
as$x=>$X){if(preg_match('~^partition~',$x))$Qg[$x]=$X;}foreach($Qg["partition_names"]as$x=>$D){if($D==""){unset($Qg["partition_names"][$x]);unset($Qg["partition_values"][$x]);}}$Qg["partition_names"]=array_values($Qg["partition_names"]);$Qg["partition_values"]=array_values($Qg["partition_values"]);if($Qg==$Sg)$Qg=array();}elseif(preg_match("~partitioned~",$S["Create_options"]))$Qg=null;$C='Table has been altered.';if($a==""){cookie("adminer_engine",$K["Engine"]);$C='Table has been created.';}$D=trim($K["name"]);$A=ME.(support("table")?"table=":"select=").url_escape($D);$I=alter_table($a,$D,(JUSH=="sqlite"&&($Kj||$nd)?$na:$m),$nd,($K["Comment"]!=$S["Comment"]?$K["Comment"]:null),($K["Engine"]&&$K["Engine"]!=$S["Engine"]?$K["Engine"]:""),($K["Collation"]&&$K["Collation"]!=$S["Collation"]?$K["Collation"]:""),($K["Auto_increment"]!=""?number($K["Auto_increment"]):""),$Qg);if($I&&!Queries::$queries)redirect($A);queries_redirect($A,$C,$I);}}page_header(($a!=""?'Alter table':'Create table'),$k,array("table"=>$a),h($a));if(!$_POST){$xj=driver()->types();$K=array("Engine"=>$_COOKIE["adminer_engine"],"fields"=>array(array("field"=>"","type"=>(isset($xj["int"])?"int":(isset($xj["integer"])?"integer":"")),"on_update"=>"")),"partition_names"=>array(""),);if($a!=""){$K=$S;$K["name"]=$a;$K["fields"]=array();if(!$_GET["auto_increment"])$K["Auto_increment"]="";foreach($_g
as$l){if($l["generated"])$l["default"]=ltrim($l["default"]);$l["generated"]=$l["generated"]?:(isset($l["default"])?"DEFAULT":"");$K["fields"][]=$l;}if($Og){$K+=$Sg;$K["partition_names"][]="";$K["partition_values"][]="";}}}$gb=collations();if(is_array(reset($gb)))$gb=call_user_func_array('array_merge',array_values($gb));$_c=driver()->engines();foreach($_c
as$zc){if(!strcasecmp($zc,$K["Engine"])){$K["Engine"]=$zc;break;}}$lf=max_input_vars(12,20);if($lf){$Qd=(count($K["fields"])>$lf?"":" hidden");echo"<p".($Qd?" id='max-fields' data-columns='$lf'":"")." class='error$Qd'>".max_input_vars_error()."\n";}echo'
<form action="" method="post" id="form">
<p>
';if(support("columns")||$a==""){echo'Table name'.": <input name='name'".($a==""&&!$_POST?" autofocus":"")." data-maxlength='64' value='".h($K["name"])."' autocapitalize='off'>\n",($_c?html_select("Engine",array(""=>"(".'engine'.")")+$_c,$K["Engine"],on('change','helpClose').on_help_value())."\n":"");if($gb)echo"<datalist id='collations'>".optionlist($gb)."</datalist>\n",(preg_match("~sqlite|mssql~",JUSH)?"":"<input list='collations' name='Collation' value='".h($K["Collation"])."' placeholder='(".'collation'.")'>\n");echo"<input type='submit' value='".'Save'."'>\n";}if(support("columns")){echo"<div class='scrollable'>\n","<table id='edit-fields' class='nowrap'>\n";edit_fields($K["fields"],$gb,"TABLE",$pd);echo"</table>\n",script("editFields();"),"</div>\n<p>\n",'Auto Increment'.": <input type='number' name='Auto_increment' class='size' value='".h($K["Auto_increment"])."'>\n",checkbox("defaults",1,($_POST?$_POST["defaults"]:get_setting("defaults")),'Default values',on('click','columnShowClick',5),"jsonly");$lb=($_POST?$_POST["comments"]:get_setting("comments"));if(support("comment")){echo
checkbox("comments",1,$lb,'Comment',on('click','editingCommentsClick',true),"jsonly").' ';$b=" name='Comment' data-maxlength='".(min_version(5.5)?2048:60)."'".($lb?"":" class='hidden'");echo
adminer()->commentInput('TABLE',$b,$K["Comment"]);}echo'<p>
<input type=\'submit\' value=\'Save\'>
';}echo'
';if($a!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$a)),'>
';if($Og&&(JUSH=='sql'||$a=="")){$Pg=preg_match('~RANGE|LIST~',$K["partition_by"]);print_fieldset("partition",'Partition by',$K["partition_by"]);echo"<p>".html_select("partition_by",array_merge(array(""),$Og),$K["partition_by"],on('change','partitionByChange').on_help_value('.','PARTITION BY $&'))."\n","(<input name='partition' value='".h($K["partition"])."'>)\n",'Partitions'.": <input type='number' name='partitions' class='size".($Pg||!$K["partition_by"]?" hidden":"")."' value='".h($K["partitions"])."'>\n","<table id='partition-table'".($Pg?"":" class='hidden'").">\n","<thead><tr><th>".'Partition name'."<th>".'Values'."<tbody>\n";foreach($K["partition_names"]as$x=>$X)echo'<tr>','<td><input name="partition_names[]" value="'.h($X).'" autocapitalize="off"'.($x==count($K["partition_names"])-1?on('input','partitionNameChange'):'').'>','<td><input name="partition_values[]" value="'.h(idx($K["partition_values"],$x)).'">';echo"</table>\n</div></fieldset>\n";}echo
input_token(),'</form>
';}elseif(isset($_GET["indexes"])){$a=$_GET["indexes"];$ke=array("PRIMARY","UNIQUE","INDEX");$S=table_status1($a,true);$ie=driver()->indexAlgorithms($S);if(preg_match('~MyISAM|M?aria'.(min_version(5.6,'10.0.5')?'|InnoDB':'').'~i',$S["Engine"]))$ke[]="FULLTEXT";if(preg_match('~MyISAM|M?aria'.(min_version(5.7,'10.2.2')?'|InnoDB':'').'~i',$S["Engine"]))$ke[]="SPATIAL";if(min_version('',11.7)&&preg_match('~MyISAM|InnoDB~i',$S["Engine"]))$ke[]="VECTOR";$w=indexes($a);$m=fields($a);$mh=array();if(JUSH=="mongo"){$mh=$w["_id_"];unset($ke[0]);unset($w["_id_"]);}$K=$_POST;if($K)save_settings(array("index_options"=>$K["options"]));if($_POST&&!$k&&!$_POST["add"]&&!$_POST["drop_col"]){$pa=array();foreach($K["indexes"]as$v){$D=$v["name"];if(in_array($v["type"],$ke)){$d=array();$Xe=array();$Xb=array();$ng=array();$je=(support("partial_indexes")?$v["partial"]:"");$he=(in_array($v["algorithm"],$ie)?$v["algorithm"]:"");$O=array();ksort($v["columns"]);foreach($v["columns"]as$x=>$c){if($c!=""){$y=idx($v["lengths"],$x);$Vb=idx($v["descs"],$x);$mg=idx($v["opclasses"],$x);$O[]=($m[$c]?idf_escape($c):$c).($y?"(".(+$y).")":"").($mg!=""?" ".idf_escape($mg):"").($Vb?" DESC":"");$d[]=$c;$Xe[]=($y?:null);$Xb[]=$Vb;$ng[]="$mg";}}$Mc=$w[$D];if($Mc){ksort($Mc["columns"]);ksort($Mc["lengths"]);ksort($Mc["descs"]);if($v["type"]==$Mc["type"]&&array_values($Mc["columns"])===$d&&(!$Mc["lengths"]||array_values($Mc["lengths"])===$Xe)&&array_values($Mc["descs"])===$Xb&&(!$Mc["opclasses"]||array_values($Mc["opclasses"])===$ng)&&$Mc["partial"]==$je&&(!$ie||$Mc["algorithm"]==$he)){unset($w[$D]);continue;}}if($d)$pa[]=array($v["type"],$D,$O,$he,$je);}}foreach($w
as$D=>$Mc)$pa[]=array($Mc["type"],$D,"DROP");if(!$pa)redirect(ME."table=".url_escape($a));queries_redirect(ME."table=".url_escape($a),'Indexes have been altered.',alter_indexes($a,$pa));}page_header('Indexes',$k,array("table"=>$a),h($a));$bd=array_keys($m);if($_POST["add"]){foreach($K["indexes"]as$x=>$v){if($v["columns"][count($v["columns"])]!="")$K["indexes"][$x]["columns"][]="";}$v=end($K["indexes"]);if($v["type"]||array_filter($v["columns"],'strlen'))$K["indexes"][]=array("columns"=>array(1=>""));}if(!$K){foreach($w
as$x=>$v){$w[$x]["name"]=$x;$w[$x]["columns"][]="";}$w[]=array("columns"=>array(1=>""));$K["indexes"]=$w;}$Xe=(JUSH=="sql"||JUSH=="mssql");$ng=driver()->indexOpclasses();$ki=($_POST?$_POST["options"]:get_setting("index_options"));echo'
<form action="" method="post">
<div class="scrollable">
<table class="nowrap odds">
<thead><tr>
<th id="label-type">Index Type
';$ae=" class='idxopts".($ki?"":" hidden")."'";if($ie)echo"<th id='label-algorithm'$ae>".'Algorithm'.doc_link(array('sql'=>'create-index.html#create-index-storage-engine-index-types','mariadb'=>'storage-engine-index-types/',));echo'<th><input type="submit" hidden>','Columns'.($Xe?"<span$ae> (".'length'.")</span>":"");if($Xe||support("descidx"))echo
checkbox("options",1,$ki,'Options',on('click','indexOptionsShow'),"jsonly")."\n";echo'<th id="label-name">Name
';if(support("partial_indexes"))echo"<th id='label-condition'$ae>".'Condition';echo'<th><noscript>',icon("plus","add[0]","+",'Add next'),'</noscript>
<tbody>
';if($mh){echo"<tr><td>PRIMARY<td>";foreach($mh["columns"]as$x=>$c)echo
select_input(" disabled",array_combine($bd,$bd),$c),"<label><input disabled type='checkbox'>".'descending'."</label> ";echo"<td><td>\n";}$De=1;foreach($K["indexes"]as$v){if(!$_POST["drop_col"]||$De!=key($_POST["drop_col"])){echo"<tr><td>".html_select("indexes[$De][type]",array(-1=>"")+$ke,$v["type"],($De==count($K["indexes"])?on('change','indexesAddRow'):""),"label-type");if($ie)echo"<td$ae>".html_select("indexes[$De][algorithm]",array_merge(array(""),$ie),$v['algorithm'],"","label-algorithm");echo"<td>";ksort($v["columns"]);$s=1;foreach($v["columns"]as$x=>$c){echo"<span>".select_input(" name='indexes[$De][columns][$s]' title='".'Column'."'".on('change','indexesChangeColumn',(JUSH=="sql"?"":$_GET["indexes"]."_")),($m&&($c==""||$m[$c])?array_combine($bd,$bd):array()),$c)," <span$ae>",($Xe?"<input type='number' name='indexes[$De][lengths][$s]' class='size' value='".h(idx($v["lengths"],$x))."' title='".'Length'."'>":"");if($ng){$mg=idx($v["opclasses"],$x);echo
html_select("indexes[$De][opclasses][$s]",array(""=>"(".'operator class'.")")+array_combine($ng,$ng)+($mg!=""?array($mg=>$mg):array()),$mg),'';}echo(support("descidx")?checkbox("indexes[$De][descs][$s]",1,idx($v["descs"],$x),'descending'):""),"<br>","</span></span>";$s++;}echo"<td><input name='indexes[$De][name]' value='".h($v["name"])."' autocapitalize='off' aria-labelledby='label-name'>\n";if(support("partial_indexes"))echo"<td$ae><input name='indexes[$De][partial]' value='".h($v["partial"])."' autocapitalize='off' aria-labelledby='label-condition'>\n";echo"<td>".icon("cross","drop_col[$De]","x",'Remove',on('click','editingRemoveRow','indexes$1[type]'));}$De++;}echo'</table>
</div>
<p>
<input type=\'submit\' value=\'Save\'>
',input_token(),'</form>
';}elseif(isset($_GET["database"])){$K=$_POST;if($_POST&&!$k&&!$_POST["add"]){$D=trim($K["name"]);if($_POST["drop"]){$_GET["db"]="";queries_redirect(remove_from_uri("db|database"),'Database has been dropped.',drop_databases(array(DB)));}elseif(DB!==$D){if(DB!=""){$_GET["db"]=$D;queries_redirect(preg_replace('~\bdb=[^&]*&~','',ME)."db=".url_escape($D),'Database has been renamed.',rename_database($D,(string)$K["collation"]));}else{$h=explode("\n",str_replace("\r","",$D));$Di=true;$Oe="";foreach($h
as$i){if(count($h)==1||$i!=""){if(!create_database($i,(string)$K["collation"]))$Di=false;$Oe=$i;}}restart_session();set_session("dbs",null);queries_redirect(ME."db=".url_escape($Oe),'Database has been created.',$Di);}}else{if(!$K["collation"])redirect(substr(ME,0,-1));query_redirect("ALTER DATABASE ".idf_escape($D).(preg_match('~^[a-z0-9_]+$~i',$K["collation"])?" COLLATE $K[collation]":""),substr(ME,0,-1),'Database has been altered.');}}page_header(DB!=""?'Alter database':'Create database',$k,array(),h(DB));$gb=collations();$D=DB;if($_POST)$D=$K["name"];elseif(DB!="")$K["collation"]=db_collation(DB,$gb);elseif(JUSH=="sql"){foreach(get_vals("SHOW GRANTS")as$_d){if(preg_match('~ ON (`(([^\\\\`]|``|\\\\.)*)%`\.\*)?~',$_d,$B)&&$B[1]){$D=stripcslashes(idf_unescape("`$B[2]`"));break;}}}echo'
<form action="" method="post">
<p>
',($_POST["add"]||strpos($D,"\n")?'<textarea autofocus name="name" rows="10" cols="40">'.h($D).'</textarea><br>':'<input name="name" autofocus value="'.h($D).'" data-maxlength="64" autocapitalize="off">')."\n",($gb?html_select("collation",array(""=>"(".'collation'.")")+$gb,$K["collation"]).doc_link(array('sql'=>"charset-charsets.html",'mariadb'=>"supported-character-sets-and-collations/",)):"")."\n",'<input type=\'submit\' value=\'Save\'>
';if(DB!="")echo"<input type='submit' name='drop' value='".'Drop'."'".confirm(sprintf('Drop %s?',DB)).">\n";elseif(!$_POST["add"]&&$_GET["db"]=="")echo
icon("plus","add[0]","+",'Add next')."\n";echo
input_token(),'</form>
';}elseif(isset($_GET["call"])){$ba=($_GET["name"]?:$_GET["call"]);page_header('Call'.": ".h($ba),$k);$Rh=(isset($_GET["callf"])?"FUNCTION":"PROCEDURE");$Ph=routine($_GET["call"],$Rh);$de=array();$Eg=array();foreach($Ph["fields"]as$s=>$l){if(substr($l["inout"],-3)=="OUT"&&JUSH=='sql')$Eg[$s]="@".idf_escape($l["field"])." AS ".idf_escape($l["field"]);if(!$l["inout"]||substr($l["inout"],0,2)=="IN")$de[]=$s;}if(!$k&&$_POST){$Qa=array();foreach($Ph["fields"]as$x=>$l){$X="";if(in_array($x,$de)){$X=process_input($l);if($X===false)$X="''";if(isset($Eg[$x]))connection()->query("SET @".idf_escape($l["field"])." = $X");}if(isset($Eg[$x]))$Qa[]="@".idf_escape($l["field"]);elseif(in_array($x,$de))$Qa[]=$X;}$H=(isset($_GET["callf"])?"SELECT ":"CALL ").(idx($Ph["returns"],"type")=="record"?"* FROM ":"").table($ba)."(".implode(", ",$Qa).")";$zi=microtime(true);$I=connection()->multi_query($H);$ka=connection()->affected_rows;echo
adminer()->selectQuery($H,$zi,!$I);if(!$I)echo"<p class='error'>".error()."\n";else{$f=connect();if($f)$f->select_db(DB);do{$I=connection()->store_result();if(is_object($I))print_select_result($I,$f);else
echo"<p class='message'>".lang_format(array('Routine has been called, %d row affected.','Routine has been called, %d rows affected.'),$ka)." <span class='time'>".@date("H:i:s")."</span>\n";}while(connection()->next_result());if($Eg)print_select_result(connection()->query("SELECT ".implode(", ",$Eg)));}}echo'
<form action="" method="post">
';if($de){echo"<table class='layout'>\n";foreach($de
as$x){$l=$Ph["fields"][$x];$D=$l["field"];echo"<tr><th>".adminer()->fieldName($l);$Y=idx($_POST["fields"],$D);if($Y!=""){if($l["type"]=="set")$Y=implode(",",$Y);}input($l,$Y,idx($_POST["function"],$D,""));echo"\n";}echo"</table>\n";}echo'<p>
<input type=\'submit\' value=\'Call\'>
',input_token(),'</form>

',adminer()->commentValue($Rh,$Ph['comment']);}elseif(isset($_GET["foreign"])){$a=$_GET["foreign"];$D=$_GET["name"];$K=$_POST;if($_POST&&!$k&&!$_POST["add"]&&!$_POST["change"]&&!$_POST["change-js"]){if(!$_POST["drop"]){$K["source"]=array_filter($K["source"],'strlen');ksort($K["source"]);$Ti=array();foreach($K["source"]as$x=>$X)$Ti[$x]=$K["target"][$x];$K["target"]=$Ti;}if(JUSH=="sqlite")$I=recreate_table($a,$a,array(),array(),array(" $D"=>($K["drop"]?"":" ".format_foreign_key($K))));else{$pa="ALTER TABLE ".table($a);$I=($D==""||queries("$pa DROP ".(JUSH=="sql"?"FOREIGN KEY ":"CONSTRAINT ").idf_escape($D)));if(!$K["drop"])$I=queries("$pa ADD".format_foreign_key($K));}queries_redirect(ME."table=".url_escape($a),($K["drop"]?'Foreign key has been dropped.':($D!=""?'Foreign key has been altered.':'Foreign key has been created.')),$I);if(!$K["drop"])$k='Source and target columns must have the same data type, there must be an index on the target columns and referenced data must exist.';}page_header(($D!=""?'Alter foreign key':'Create foreign key'),$k,array("table"=>$a),h($D!=""?$D:$a));if($_POST){ksort($K["source"]);if($_POST["change"]||$_POST["change-js"])$K["target"]=array();else$K["source"][]="";}elseif($D!=""){$pd=foreign_keys($a);$K=$pd[$D];$K["source"][]="";}else{$K["table"]=$a;$K["source"]=array("");}echo'
<form action="" method="post">
';$ri=array_keys(fields($a));if($K["db"]!="")connection()->select_db($K["db"]);if($K["ns"]!=""){$Ag=get_schema();set_schema($K["ns"]);}$Ah=array_keys(array_filter(table_status('',true),'Adminer\fk_support'));$Ti=array_keys(fields(in_array($K["table"],$Ah)?$K["table"]:reset($Ah)));$b=on('change','foreignChange');echo"<p><label>".'Target table'.": ".html_select("table",$Ah,$K["table"],$b)."</label>\n";if(JUSH!="sqlite"){$Nb=array();foreach(adminer()->databases()as$i){if(!information_schema($i))$Nb[]=$i;}echo"<label>".'DB'.": ".html_select("db",$Nb,$K["db"]!=""?$K["db"]:$_GET["db"],$b)."</label>";}echo
input_hidden("change-js"),'<noscript><p><input type=\'submit\' name=\'change\' value=\'Change\'></noscript>
<table>
<thead><tr><th id="label-source">Source<th id="label-target">Target<tbody>
';$De=0;foreach($K["source"]as$x=>$X){echo"<tr>","<td>".html_select("source[".(+$x)."]",array(-1=>"")+$ri,$X,($De==count($K["source"])-1?on('change','foreignAddRow'):""),"label-source"),"<td>".html_select("target[".(+$x)."]",$Ti,idx($K["target"],$x),"","label-target");$De++;}echo'</table>
<p>
<label>ON DELETE: ',html_select("on_delete",array(-1=>"")+explode("|",driver()->onActions),$K["on_delete"]),'</label>
<label>ON UPDATE: ',html_select("on_update",array(-1=>"")+explode("|",driver()->onActions),$K["on_update"]),'</label>
',(support("deferrable")?html_select("deferrable",array('NOT DEFERRABLE','DEFERRABLE','DEFERRABLE INITIALLY DEFERRED'),$K["deferrable"]).' ':''),doc_link(array('sql'=>"innodb-foreign-key-constraints.html",'mariadb'=>"foreign-keys/",)),'<p>
<input type=\'submit\' value=\'Save\'>
<noscript><p><input type=\'submit\' name=\'add\' value=\'Add column\'></noscript>
';if($D!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$D)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["view"])){$a=$_GET["view"];$K=$_POST;$Bg="VIEW";if(JUSH=="pgsql"&&$a!=""){$P=table_status1($a);$Bg=strtoupper($P["Engine"]);}if($_POST&&!$k){$D=trim($K["name"]);$va=" AS\n$K[select]";$A=ME."table=".url_escape($D);$C='View has been altered.';$U=($_POST["materialized"]?"MATERIALIZED VIEW":"VIEW");if(!$_POST["drop"]&&$a==$D&&JUSH!="sqlite"&&$U=="VIEW"&&$Bg=="VIEW")query_redirect((JUSH=="mssql"?"ALTER":"CREATE OR REPLACE")." VIEW ".table($D).$va,$A,$C);else{$Vi="adminer_".uniqid();drop_create("DROP $Bg ".table($a),"CREATE $U ".table($D).$va,"DROP $U ".table($D),"CREATE $U ".table($Vi).$va,"DROP $U ".table($Vi),($_POST["drop"]?substr(ME,0,-1):$A),'View has been dropped.',$C,'View has been created.',$a,$D);}}if(!$_POST&&$a!=""){$K=view($a);$K["name"]=$a;$K["materialized"]=($Bg!="VIEW");if(!$k)$k=error();}page_header(($a!=""?'Alter view':'Create view'),$k,array("table"=>$a),h($a));echo'
<form action="" method="post">
<p>Name: <input name="name" value="',h($K["name"]),'" data-maxlength="64" autocapitalize="off">
',(support("materializedview")?" ".checkbox("materialized",1,$K["materialized"],'Materialized view'):""),'<p>';textarea("select",$K["select"]);echo'<p>
<input type=\'submit\' value=\'Save\'>
';if($a!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$a)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["event"])){$aa=$_GET["event"];$ue=array("YEAR","QUARTER","MONTH","DAY","HOUR","MINUTE","WEEK","SECOND","YEAR_MONTH","DAY_HOUR","DAY_MINUTE","DAY_SECOND","HOUR_MINUTE","HOUR_SECOND","MINUTE_SECOND");$_i=array("ENABLED"=>"ENABLE","DISABLED"=>"DISABLE","SLAVESIDE_DISABLED"=>"DISABLE ON SLAVE");$K=$_POST;if($_POST&&!$k){if($_POST["drop"])query_redirect("DROP EVENT ".idf_escape($aa),substr(ME,0,-1),'Event has been dropped.');elseif(in_array($K["INTERVAL_FIELD"],$ue)&&isset($_i[$K["STATUS"]])){$Uh="\nON SCHEDULE ".($K["INTERVAL_VALUE"]?"EVERY ".q($K["INTERVAL_VALUE"])." $K[INTERVAL_FIELD]".($K["STARTS"]?" STARTS ".q($K["STARTS"]):"").($K["ENDS"]?" ENDS ".q($K["ENDS"]):""):"AT ".q($K["STARTS"]))." ON COMPLETION".($K["ON_COMPLETION"]?"":" NOT")." PRESERVE";queries_redirect(substr(ME,0,-1),($aa!=""?'Event has been altered.':'Event has been created.'),queries(($aa!=""?"ALTER EVENT ".idf_escape($aa).$Uh.($aa!=$K["EVENT_NAME"]?"\nRENAME TO ".idf_escape($K["EVENT_NAME"]):""):"CREATE EVENT ".idf_escape($K["EVENT_NAME"]).$Uh)."\n".$_i[$K["STATUS"]]." COMMENT ".q($K["EVENT_COMMENT"]).rtrim(" DO\n$K[EVENT_DEFINITION]",";").";"));}}page_header(($aa!=""?'Alter event'.": ".h($aa):'Create event'),$k);if(!$K&&$aa!=""){$L=get_rows("SELECT * FROM information_schema.EVENTS WHERE EVENT_SCHEMA = ".q(DB)." AND EVENT_NAME = ".q($aa));$K=reset($L);}echo'
<form action="" method="post">
<table class="layout">
<tr><th>Name<td><input name="EVENT_NAME" value="',h($K["EVENT_NAME"]),'" data-maxlength="64" autocapitalize="off">
<tr><th title="datetime">Start<td><input name="STARTS" value="',h("$K[EXECUTE_AT]$K[STARTS]"),'">
<tr><th title="datetime">End<td><input name="ENDS" value="',h($K["ENDS"]),'">
<tr><th>Every
<td><input type="number" name="INTERVAL_VALUE" value="',h($K["INTERVAL_VALUE"]),'" class="size"> ',html_select("INTERVAL_FIELD",$ue,$K["INTERVAL_FIELD"]),'<tr><th>Status<td>',html_select("STATUS",$_i,$K["STATUS"]),'<tr><th>Comment<td><input name="EVENT_COMMENT" value="',h($K["EVENT_COMMENT"]),'" data-maxlength="64">
<tr><th><td>',checkbox("ON_COMPLETION","PRESERVE",$K["ON_COMPLETION"]=="PRESERVE",'On completion preserve'),'</table>
<p>';textarea("EVENT_DEFINITION",$K["EVENT_DEFINITION"]);echo'<p>
<input type=\'submit\' value=\'Save\'>
';if($aa!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$aa)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["procedure"])){$ba=($_GET["name"]?:$_GET["procedure"]);$Ph=(isset($_GET["function"])?"FUNCTION":"PROCEDURE");$K=$_POST;$K["fields"]=(array)$K["fields"];if($_POST&&!process_fields($K["fields"])&&!$k){foreach($K["fields"]as$x=>$l){if($l["field"]=="")unset($K["fields"][$x]);}$gg=routine_id($ba,routine($_GET["procedure"],$Ph));$Qf=routine_id($K["name"],$K);$g=create_routine($Ph,$K);$A=substr(ME,0,-1);$C='Routine has been altered.';if(!$_POST["drop"]&&$gg==$Qf&&connection()->flavor!="mysql")query_redirect(substr_replace($g,' OR REPLACE',6,0),$A,$C);else{$Vi="adminer_".uniqid();drop_create("DROP $Ph $gg",$g,"DROP $Ph $Qf",create_routine($Ph,array("name"=>$Vi)+$K),"DROP $Ph ".routine_id($Vi,$K),$A,'Routine has been dropped.',$C,'Routine has been created.',$ba,$K["name"]);}}page_header(($ba!=""?(isset($_GET["function"])?'Alter function':'Alter procedure').": ".h($ba):(isset($_GET["function"])?'Create function':'Create procedure')),$k);if(!$_POST){if($ba=="")$K["language"]="sql";else{$K=routine($_GET["procedure"],$Ph);$K["name"]=$ba;}}$gb=get_vals("SHOW CHARACTER SET");sort($gb);$Qh=routine_languages();echo($gb?"<datalist id='collations'>".optionlist($gb)."</datalist>":""),'
<form action="" method="post" id="form">
<p>Name: <input name="name" value="',h($K["name"]),'" data-maxlength="64" autocapitalize="off">
',($Qh?"<label>".'Language'.": ".html_select("language",$Qh,$K["language"])."</label>\n":""),'<input type=\'submit\' value=\'Save\'>
<div class="scrollable">
<table id="edit-fields" class="nowrap">
';edit_fields($K["fields"],$gb,$Ph);if(isset($_GET["function"])){echo"<tr><td>".'Return type';edit_type("returns",(array)$K["returns"],$gb,array(),(JUSH=="pgsql"?array("void","trigger"):array()));}echo'</table>
',script("editFields();"),'</div>
<p>';textarea("definition",$K["definition"],20);echo'<p>
<input type=\'submit\' value=\'Save\'>
';if($ba!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$ba)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["check"])){$a=$_GET["check"];$D=$_GET["name"];$K=$_POST;if($K&&!$k){if(JUSH=="sqlite")$I=recreate_table($a,$a,array(),array(),array(),"",array(),"$D",($K["drop"]?"":$K["clause"]));else{$I=($D==""||queries("ALTER TABLE ".table($a)." DROP CONSTRAINT ".idf_escape($D)));if(!$K["drop"])$I=queries("ALTER TABLE ".table($a)." ADD".($K["name"]!=""?" CONSTRAINT ".idf_escape($K["name"]):"")." CHECK ($K[clause])");}queries_redirect(ME."table=".url_escape($a),($K["drop"]?'Check has been dropped.':($D!=""?'Check has been altered.':'Check has been created.')),$I);}page_header(($D!=""?'Alter check':'Create check'),$k,array("table"=>$a),h($D!=""?$D:$a));if(!$K){$Xa=driver()->checkConstraints($a);$K=array("name"=>$D,"clause"=>$Xa[$D]);}echo'
<form action="" method="post">
<p>';if(JUSH!="sqlite")echo'Name'.': <input name="name" value="'.h($K["name"]).'" data-maxlength="64" autocapitalize="off"> ';echo
doc_link(array('sql'=>"create-table-check-constraints.html",'mariadb'=>"constraint/",),"?"),'<p>';textarea("clause",$K["clause"]);echo'<p><input type=\'submit\' value=\'Save\'>
';if($D!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$D)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["trigger"])){$a=$_GET["trigger"];$D="$_GET[name]";$rj=trigger_options();$K=(array)trigger($D,$a)+array("Trigger"=>$a."_bi");if($_POST){if(!$k&&in_array($_POST["Timing"],$rj["Timing"])&&in_array($_POST["Event"],$rj["Event"])&&in_array($_POST["Type"],$rj["Type"])){$jg=" ON ".table($a);$nc="DROP TRIGGER ".idf_escape($D).(JUSH=="pgsql"?$jg:"");$A=ME."table=".url_escape($a);if($_POST["drop"])query_redirect($nc,$A,'Trigger has been dropped.');else{if($D!="")queries($nc);queries_redirect($A,($D!=""?'Trigger has been altered.':'Trigger has been created.'),queries(create_trigger($jg,$_POST)));if($D!="")queries(create_trigger($jg,$K+array("Type"=>reset($rj["Type"]))));}}$K=$_POST;}page_header(($D!=""?'Alter trigger':'Create trigger'),$k,array("table"=>$a),h($D!=""?$D:$a));$qj=on('change','triggerChange',"^".preg_quote($a,"/")."_[ba][iud]$",$a);echo'
<form action="" method="post" id="form">
<table class="layout">
<tr><th>Time
<td>',html_select("Timing",$rj["Timing"],$K["Timing"],$qj),'<tr><th>Event<td>',html_select("Event",$rj["Event"],$K["Event"],$qj),(in_array("UPDATE OF",$rj["Event"])?" <input name='Of' value='".h($K["Of"])."' class='hidden'>":""),'<tr><th>Type<td>',html_select("Type",$rj["Type"],$K["Type"]),'</table>
<p>Name: <input name="Trigger" value="',h($K["Trigger"]),'" data-maxlength="64" autocapitalize="off">
',script("fire(qs('#form')['Timing'], 'change');"),'<p>';textarea("Statement",$K["Statement"]);echo'<p>
<input type=\'submit\' value=\'Save\'>
';if($D!="")echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',$D)),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["user"])){$da=$_GET["user"];$qh=array(""=>array("All privileges"=>""));foreach(get_rows("SHOW PRIVILEGES")as$K){foreach(explode(",",($K["Privilege"]=="Grant option"?"":$K["Context"]))as$tb)$qh[$tb=="File access on server"?"Server Admin":$tb][$K["Privilege"]]=$K["Comment"];}unset($qh["Server Admin"]["Usage"]);foreach($qh["Tables"]as$x=>$X)unset($qh["Databases"][$x]);$Pf=array();if($_POST){foreach($_POST["objects"]as$x=>$X)$Pf[$X]=(array)$Pf[$X]+idx($_POST["grants"],$x,array());}$Ad=array();if(isset($_GET["host"])&&($I=connection()->query("SHOW GRANTS FOR ".q($da)."@".q($_GET["host"])))){while($K=$I->fetch_row()){if(preg_match('~GRANT (.*) ON (.*) TO ~',$K[0],$B)&&preg_match_all('~ *([^(,]*[^ ,(])( *\([^)]+\))?~',$B[1],$if,PREG_SET_ORDER)){foreach($if
as$X){if($X[1]!="USAGE")$Ad["$B[2]$X[2]"][$X[1]]=true;if(preg_match('~ WITH GRANT OPTION~',$K[0]))$Ad["$B[2]$X[2]"]["GRANT OPTION"]=true;}}}}if($_POST&&!$k){$ig=(isset($_GET["host"])?q($da)."@".q($_GET["host"]):"''");if($_POST["drop"])query_redirect("DROP USER $ig",ME."privileges=",'User has been dropped.');else{$Sf=q($_POST["user"])."@".q($_POST["host"]);$Tg=$_POST["pass"];$_b=false;$I=true;if($ig!=$Sf){$_b=queries("CREATE USER $Sf IDENTIFIED BY ".($_POST["hashed"]?"PASSWORD ":"").q($Tg));$I=$_b;}elseif($Tg!="")$I=queries("SET PASSWORD FOR $Sf = ".(min_version(8,99)||$_POST["hashed"]?q($Tg):"PASSWORD(".q($Tg).")"));if($I){$Mh=array();foreach($Pf
as$Zf=>$_d){if(isset($_GET["grant"]))$_d=array_filter($_d);$_d=array_keys($_d);if(isset($_GET["grant"]))$Mh=array_diff(array_keys(array_filter($Pf[$Zf],'strlen')),$_d);elseif($ig==$Sf){$fg=array_keys((array)$Ad[$Zf]);$Mh=array_diff($fg,$_d);$_d=array_diff($_d,$fg);unset($Ad[$Zf]);}if(preg_match('~^(.+)\s*(\(.*\))?$~U',$Zf,$B)&&(!grant("REVOKE",$Mh,$B[2]," ON $B[1] FROM $Sf")||!grant("GRANT",$_d,$B[2]," ON $B[1] TO $Sf"))){$I=false;break;}}}if($I&&isset($_GET["host"])){if($ig!=$Sf)queries("DROP USER $ig");elseif(!isset($_GET["grant"])){foreach($Ad
as$Zf=>$Mh){if(preg_match('~^(.+)(\(.*\))?$~U',$Zf,$B))grant("REVOKE",array_keys($Mh),$B[2]," ON $B[1] FROM $Sf");}}}if($I&&!Queries::$queries)redirect(ME."privileges=");queries_redirect(ME."privileges=",(isset($_GET["host"])?'User has been altered.':'User has been created.'),$I);if($_b)connection()->query("DROP USER $Sf");}}page_header((isset($_GET["host"])?'Username'.": ".h("$da@$_GET[host]"):'Create user'),$k,array("privileges"=>array('','Privileges')));$K=$_POST;if($K)$Ad=$Pf;else{$K=$_GET+array("host"=>get_val("SELECT SUBSTRING_INDEX(CURRENT_USER, '@', -1)"));$Ad[(DB==""||$Ad?"":idf_escape(addcslashes(DB,"%_\\"))).".*"]=array();}echo'<form action="" method="post">
<table class="layout">
<tr><th>Server<td><input name="host" data-maxlength="60" value="',h($K["host"]),'" autocapitalize="off">
<tr><th>Username<td><input name="user" data-maxlength="80" value="',h($K["user"]),'" autocapitalize="off">
<tr><th>Password<td><input name="pass" id="pass" value="',h($K["pass"]),'" autocomplete="new-password">
',($K["hashed"]?"":script("typePassword(qs('#pass'));")),(min_version(8,99)?"":checkbox("hashed",1,$K["hashed"],'Hashed',on('click','hashedClick'))),'</table>

',"<table class='odds'>\n","<thead><tr><th colspan='2'>".'Privileges'.doc_link(array('sql'=>"grant.html#priv_level"));$s=0;foreach($Ad
as$Zf=>$_d){echo'<th>'.($Zf!="*.*"?"<input name='objects[$s]' value='".h($Zf)."' size='10' autocapitalize='off'>":input_hidden("objects[$s]","*.*")."*.*");$s++;}echo"<tbody>\n";foreach(array(""=>"","Server Admin"=>'Server',"Databases"=>'Database',"Tables"=>'Table',"Procedures"=>'Routine',)as$tb=>$Vb){foreach((array)$qh[$tb]as$ph=>$jb){echo"<tr><td".($Vb?">$Vb<td":" colspan='2'").' lang="en" title="'.h($jb).'">'.h($ph);$s=0;foreach($Ad
as$Zf=>$_d){$D="'grants[$s][".h(strtoupper($ph))."]'";$Y=$_d[strtoupper($ph)];if($tb=="Server Admin"&&$Zf!=(isset($Ad["*.*"])?"*.*":".*"))echo"<td>";elseif(isset($_GET["grant"]))echo"<td><select name=$D><option><option value='1'".($Y?" selected":"").">".'Grant'."<option value='0'".($Y=="0"?" selected":"").">".'Revoke'."</select>";else
echo"<td align='center'><label class='block'>","<input type='checkbox' name=$D value='1'".($Y?" checked":"").($ph=="All privileges"?" id='grants-$s-all'":($ph=="Grant option"?"":on('click','grantsClick',"grants-$s-all"))).">","</label>";$s++;}}}echo"</table>\n",'<p>
<input type=\'submit\' value=\'Save\'>
';if(isset($_GET["host"]))echo'<input type=\'submit\' name=\'drop\' value=\'Drop\'',confirm(sprintf('Drop %s?',"$da@$_GET[host]")),'>
';echo
input_token(),'</form>
';}elseif(isset($_GET["processlist"])){if(support("kill")){if($_POST&&!$k){$Ke=0;foreach((array)$_POST["kill"]as$X){if(adminer()->killProcess($X))$Ke++;}queries_redirect(ME."processlist=",lang_format(array('%d process has been killed.','%d processes have been killed.'),$Ke),$Ke||!$_POST["kill"]);}}page_header('Process list',$k);echo'
<form action="" method="post">
<div class="scrollable">
<table class="nowrap checkable odds"',on('click','tableClick').on('dblclick','tableClick'),'>
';$s=-1;foreach(adminer()->processList()as$s=>$K){if(!$s){echo"<thead><tr lang='en'>".(support("kill")?"<td class='hover'>":"");foreach($K
as$x=>$X)echo"<th>$x".doc_link(array('sql'=>"show-processlist.html#processlist_".strtolower($x),));echo"<tbody>\n";}echo"<tr>".(support("kill")?"<td class='hover'>".checkbox("kill[]",$K[JUSH=="sql"?"Id":"pid"],0):"");foreach($K
as$x=>$X)echo"<td>".($X!=""&&((JUSH=="sql"&&$x=="Info"&&preg_match("~Query|Killed~",$K["Command"]))||(JUSH=="pgsql"&&$x=="query")||(JUSH=="oracle"&&$x=="sql_text"))?"<code class='jush-".JUSH."' data-full='".h($X)."'>".shorten_utf8($X,100,"</code>").' <a href="'.h(ME.($K["db"]!=""?"db=".url_escape($K["db"])."&":"")."sql=".url_escape($X)).'">'.'Clone'.'</a>'.' '.copy_icon():h($X));echo"\n";}echo'</table>
</div>
<p>
',script("copyCode(qsl('table'));");if(support("kill"))echo($s+1)."/".sprintf('%d in total',max_connections()),"<p><input type='submit' value='".'Kill'."'>\n";echo
input_token(),'</form>
',script("tableCheck();");}elseif(isset($_GET["select"])){$a=$_GET["select"];$S=table_status1($a);$w=indexes($a);$m=fields($a);$pd=column_foreign_keys($a);$eg=$S["Oid"];$ja=get_settings("adminer_import");$Nh=array();$d=array();$Yh=array();$ug=array();$Yi=null;foreach($m
as$x=>$l){$D=adminer()->fieldName($l);$Nf=html_entity_decode(strip_tags($D),ENT_QUOTES);if(isset($l["privileges"]["select"])&&$D!=""){$d[$x]=$Nf;if(is_shortable($l))$Yi=adminer()->selectLengthProcess();}if(isset($l["privileges"]["where"])&&$D!="")$Yh[$x]=$Nf;if(isset($l["privileges"]["order"])&&$D!="")$ug[$x]=$Nf;$Nh+=$l["privileges"];}list($M,$r)=adminer()->selectColumnsProcess($d,$w);$M=array_unique($M);$r=array_unique($r);$ze=count($r)<count($M);$Z=adminer()->selectSearchProcess($m,$w);$E=adminer()->selectOrderProcess($m,$w);$z=adminer()->selectLimitProcess();if($_GET["val"]&&is_ajax()){header("Content-Type: text/plain; charset=utf-8");foreach($_GET["val"]as$Aj=>$K){$va=convert_field($m[key($K)]);$M=array($va?:idf_escape(key($K)));$Z[]=where_check(bracket_escape($Aj,true),$m);$J=driver()->select($a,$M,$Z,$M);if($J)echo
first($J->fetch_row());}exit;}$mh=$Cj=array();foreach($w
as$v){if($v["type"]=="PRIMARY"){$mh=array_flip($v["columns"]);$Cj=($M?$mh:array());foreach($Cj
as$x=>$X){if(in_array(idf_escape($x),$M))unset($Cj[$x]);}break;}}if($eg&&!$mh){$mh=$Cj=array($eg=>0);$w[]=array("type"=>"PRIMARY","columns"=>array($eg));}if($_POST&&!$k){$ek=$Z;if(!$_POST["all"]&&is_array($_POST["check"])){$Xa=array();foreach($_POST["check"]as$Ua)$Xa[]=where_check($Ua,$m);$ek[]="((".implode(") OR (",$Xa)."))";}$gk=$ek;$ek=($ek?"\nWHERE ".implode(" AND ",$ek):"");if($_POST["export"]){save_settings(array("output"=>$_POST["output"],"format"=>$_POST["format"]),"adminer_import");dump_headers($a);adminer()->dumpTable($a,"");$ai=($M?:array("*"));$vb=convert_fields($d,$m,$M);if($vb)$ai[]=substr($vb,2);$H="";if(is_array($_POST["check"])&&!$mh){$td=implode(", ",$ai)."\nFROM ".table($a);$Cd=($r&&$ze?"\nGROUP BY ".implode(", ",$r):"").($E?"\nORDER BY ".implode(", ",$E):"");$zj=array();foreach($_POST["check"]as$X)$zj[]="(SELECT".limit($td,"\nWHERE ".($Z?implode(" AND ",$Z)." AND ":"").where_check($X,$m).$Cd,1).")";$H=implode(" UNION ALL ",$zj);}adminer()->dumpData($a,"table",$H,$ai,$gk,($ze?$r:array()),$E);adminer()->dumpFooter();exit;}if(!adminer()->selectEmailProcess($Z,$pd)){if($_POST["save"]||$_POST["delete"]){$I=true;$ka=0;$O=array();if(!$_POST["delete"]){foreach($m
as$D=>$X){$u=bracket_escape($D);if(isset($_POST["fields"][$u])||$_FILES["fields-$u"]){$X=process_input($m[$D]);if($X!==null&&($_POST["clone"]||$X!==false))$O[idf_escape($D)]=($X!==false?$X:idf_escape($D));}}}if($_POST["delete"]||$O){$H=($_POST["clone"]?"INTO ".table($a)." (".implode(", ",array_keys($O)).")\nSELECT ".implode(", ",$O)."\nFROM ".table($a):"");if($_POST["all"]||($mh&&is_array($_POST["check"]))||$ze){$I=($_POST["delete"]?driver()->delete($a,$ek):($_POST["clone"]?queries("INSERT $H$ek".driver()->insertReturning($a)):driver()->update($a,$O,$ek)));$ka=connection()->affected_rows;if(is_object($I))$ka+=$I->num_rows;}else{foreach((array)$_POST["check"]as$X){$dk="\nWHERE ".($Z?implode(" AND ",$Z)." AND ":"").where_check($X,$m);$I=($_POST["delete"]?driver()->delete($a,$dk,1):($_POST["clone"]?queries("INSERT".limit1($a,$H,$dk)):driver()->update($a,$O,$dk,1)));if(!$I)break;$ka+=connection()->affected_rows;}}}$C=lang_format(array('%d item has been affected.','%d items have been affected.'),$ka);if($_POST["clone"]&&$I&&$ka==1){$Qe=last_id($I);if($Qe)$C=sprintf('Item%s has been inserted.'," $Qe");}queries_redirect(remove_from_uri($_POST["all"]&&$_POST["delete"]?"page|next":""),$C,$I);if(!$_POST["delete"]){$gh=(array)$_POST["fields"];edit_form($a,array_intersect_key($m,$gh),$gh,!$_POST["clone"],$k);page_footer();exit;}}elseif(!$_POST["import"]){$I=true;$ka=0;foreach((array)$_POST["val"]as$Aj=>$K){$O=array();foreach($K
as$x=>$X){$x=bracket_escape($x,true);$O[idf_escape($x)]=(preg_match('~char|text~',$m[$x]["type"])||$X!=""?adminer()->processInput($m[$x],$X):"NULL");}$I=driver()->update($a,$O," WHERE ".($Z?implode(" AND ",$Z)." AND ":"").where_check(bracket_escape($Aj,true),$m),($ze||$mh?0:1)," ");if(!$I)break;$ka+=connection()->affected_rows;}queries_redirect(remove_from_uri(),lang_format(array('%d item has been affected.','%d items have been affected.'),$ka),$I);}elseif(!is_string($cd=get_file("csv_file",true)))$k=upload_error($cd);elseif(!preg_match('~~u',$cd))$k='File must be in UTF-8 encoding.';else{save_settings(array("output"=>$ja["output"],"format"=>$_POST["separator"]),"adminer_import");$hb=array_keys($m);$ei=($_POST["separator"]=="csv"?",":($_POST["separator"]=="tsv"?"\t":";"));$Db=parse_csv($cd,$ei);$ka=count($Db);driver()->begin();$L=array();foreach($Db
as$x=>$Tj){if(!$x&&!array_diff($Tj,$hb)){$hb=$Tj;$ka--;}else{$O=array();foreach($Tj
as$s=>$db)$O[idf_escape($hb[$s])]=($db==""&&$m[$hb[$s]]["null"]?"NULL":q(csv_value($db)));$L[]=$O;}}$I=(!$L||driver()->insertUpdate($a,$L,$mh));if($I)driver()->commit();queries_redirect(remove_from_uri("page|next"),lang_format(array('%d row has been imported.','%d rows have been imported.'),$ka),$I);driver()->rollback();}}}$Li=adminer()->tableName($S);if(is_ajax()){page_headers();ob_start();}else
page_header('Select'.": $Li",$k);$O=null;if(isset($Nh["insert"])||!support("table")){$O="";foreach((array)$_GET["where"]as$X){$Y=$X["val"];if(is_array($Y))$Y=(count($Y)==1&&preg_match('~^val-(.*)~s',reset($Y),$B)?$B[1]:"");if($X["col"]!=""&&$Y!=""&&($X["op"]=="="||(!$X["op"]&&(is_array($X["val"])||!preg_match('~[_%]~',$Y)))))$O
.="&set[".url_escape(bracket_escape($X["col"]))."]=".url_escape($Y);}}adminer()->selectLinks($S,$O);if(!$d&&support("table"))echo"<p class='error'>".'Unable to select the table'.($m?".":": ".error())."\n";else{echo"<form action='' id='form'>\n","<div hidden>";hidden_fields_get();echo(DB!=""?input_hidden("db",DB).(isset($_GET["ns"])?input_hidden("ns",$_GET["ns"]):""):""),input_hidden("select",$a),"</div>\n";adminer()->selectColumnsPrint($M,$d);adminer()->selectSearchPrint($Z,$Yh,$w);adminer()->selectOrderPrint($E,$ug,$w);adminer()->selectLimitPrint($z);if($Yi!==null)adminer()->selectLengthPrint($Yi);adminer()->selectActionPrint($w);echo"</form>\n";foreach((array)$_GET["where"]as$X){if($X["op"]=="SQL"&&!in_array($_SERVER["HTTP_SEC_FETCH_SITE"],array("","same-origin"))){echo"<p class='error'>".'Invalid CSRF token. Send the form again.'.' '.'If you did not send this request from Adminer then close this page.'."\n";page_footer();exit;}}$F=$_GET["page"];$sd=null;if($F=="last"){$sd=get_val(count_rows($a,$Z,$ze,$r));$F=floor(max(0,intval($sd)-1)/$z);}$Zh=$M;$Bd=$r;if(!$Zh){$Zh[]="*";$vb=convert_fields($d,$m,$M);if($vb)$Zh[]=substr($vb,2);}foreach($M
as$x=>$X){$l=$m[idf_unescape($X)];if($l&&($va=convert_field($l)))$Zh[$x]="$va AS $X";}if(JUSH=="pgsql"||JUSH=="mssql"){foreach((array)$_GET["columns"]as$x=>$X){if(isset($Zh[$x])&&$X["fun"])$Zh[$x].=" AS ".idf_escape(apply_sql_function($X["fun"],($X["col"]!=""?$X["col"]:"*")));}}if(!$ze&&$Cj){foreach($Cj
as$x=>$X){$Zh[]=idf_escape($x);if($Bd)$Bd[]=idf_escape($x);}}$I=driver()->select($a,$Zh,$Z,$Bd,$E,$z,$F,true);if(!is_object($I))echo"<p class='error'>".(error()?:'Unknown error.')."\n";else{if(JUSH=="mssql"&&$F)$I->seek($z*$F);$xc=array();$L=array();while($K=$I->fetch_assoc()){if($F&&JUSH=="oracle")unset($K["RNUM"]);$L[]=$K;}$Kd=($z&&(support("cursor")?$_GET["next"]!="":count($L)>=$z));if(is_ajax()&&$Kd)header("X-Next-Page: ".pagination_href($F+1));if($_GET["modify"]&&$L){$rf=max_input_vars(count($L[0])+1,20);echo($rf&&count($L)>$rf?"<p class='error'>".max_input_vars_error()."\n":"");}echo"<form action='' method='post' enctype='multipart/form-data'>\n";if($_GET["page"]!="last"&&$z&&$r&&$ze&&JUSH=="sql")$sd=get_val(" SELECT FOUND_ROWS()");if(!$L)echo"<p class='message'>".'No rows.'."\n";else{$Ea=adminer()->backwardKeys($a,$Li);echo"<div class='scrollable'>","<table id='table' class='nowrap checkable odds'".on('click','tableClick').on('dblclick','tableClick').on('keydown','editingKeydown').">\n","<thead><tr>".(!$r&&$M?"":"<td class='hover check'><input type='checkbox' id='all-page' class='jsonly' title='".'All rows on this page'."'".on('click','formCheck','^check').">");$Of=array();$xd=array();reset($M);$yh=1;foreach($L[0]as$x=>$X){if(!isset($Cj[$x])){$X=idx($_GET["columns"],key($M))?:array();$l=$m[$M?($X?$X["col"]:current($M)):$x];$D=($l?adminer()->fieldName($l,$yh):($X["fun"]?"*":h($x)));if($D!=""){$yh++;$Of[$x]=$D;$c=idf_escape($x);$Vd=remove_from_uri('(order|desc)[^=]*|page|next').'&order[0]='.url_escape($x);$Vb="&desc[0]=1";$oi=preg_replace('~ DESC( NULLS LAST)?$~','',$E[0]);$qi=($oi==$c||$oi==$x);echo"<th id='th[".h(bracket_escape($x))."]'".($qi?" aria-sort='".($oi==$E[0]?"ascending":"descending")."'":"").">";$wd=apply_sql_function($X["fun"],$D);$pi=isset($l["privileges"]["order"])||$wd!=$D;echo($pi?"<a href='".h($Vd.($qi&&$oi==$E[0]?$Vb:''))."'>$wd</a>":$wd);$yf=($pi?"<a href='".h($Vd.$Vb)."' title='".'descending'."' class='text'> ↓</a>":'');if(!$X["fun"]&&isset($l["privileges"]["where"]))$yf
.="<a href='#fieldset-search' title='".'Search'."' class='text jsonly'".on('click','selectSearch',$x)."> =</a>";echo($yf?"<span class='column'>$yf</span>":"");}$xd[$x]=$X["fun"];next($M);}}$Xe=array();if($_GET["modify"]){foreach($L
as$K){foreach($K
as$x=>$X)$Xe[$x]=max($Xe[$x],min(40,strlen(utf8_decode($X))));}}echo($Ea?"<th>".'Relations':"")."<tbody>\n";if(is_ajax())ob_end_clean();foreach(adminer()->rowDescriptions($L,$pd)as$Mf=>$K){$_j=unique_array($L[$Mf],$w);if(!$_j){$_j=array();reset($M);foreach($L[$Mf]as$x=>$X){if(!preg_match('~^(COUNT|AVG|GROUP_CONCAT|MAX|MIN|SUM)\(~',current($M)))$_j[$x]=$X;next($M);}}$Aj="";foreach($_j
as$x=>$X){$l=(array)$m[$x];$ye=is_blob($l);if((JUSH=="sql"||JUSH=="pgsql")&&($ye||preg_match('~char|text|enum|set~',$l["type"]))&&strlen($X)>64){$x=(strpos($x,'(')?$x:idf_escape($x));$x="MD5(".($ye||JUSH!='sql'||preg_match("~^utf8~",$l["collation"])?$x:"CONVERT($x USING ".charset(connection()).")").")";$X=md5($ye?(string)driver()->value($X,$l):$X);}$Aj
.="&".($X!==null?"where[".url_escape(bracket_escape($x))."]=".url_escape($X===false?"f":$X):"null[]=".url_escape($x));}echo"<tr>".(!$r&&$M?"":"<td class='hover check'>".($ze||information_schema(DB)?"":"<a href='".h(ME."edit=".url_escape($a).$Aj)."' class='edit'>".'edit'."</a> ").checkbox("check[]",substr($Aj,1),in_array(substr($Aj,1),(array)$_POST["check"])));reset($M);foreach($K
as$x=>$X){if(isset($Of[$x])){$c=current($M);$l=(array)$m[$x];if($X!=""&&(!isset($xc[$x])||$xc[$x]!=""))$xc[$x]=(is_mail($X)?$Of[$x]:"");$_="";if(is_blob($l)&&$X!="")$_=ME.'download='.url_escape($a).'&field='.url_escape($x).$Aj;if(!$_&&$X!==null){foreach((array)$pd[$x]as$o){if(count($pd[$x])==1||end($o["source"])==$x){$_="";foreach($o["source"]as$s=>$ri)$_
.=where_link($s,$o["target"][$s],$L[$Mf][$ri]);$_=($o["db"]!=""?preg_replace('~([?&]db=)[^&]+~','\1'.url_escape($o["db"]),ME):ME).'select='.url_escape($o["table"]).$_;if($o["ns"])$_=preg_replace('~([?&]ns=)[^&]+~','\1'.url_escape($o["ns"]),$_);if(count($o["source"])==1)break;}}}if($c=="COUNT(*)"){$_=ME."select=".url_escape($a);$s=0;foreach((array)$_GET["where"]as$W){if(!array_key_exists($W["col"],$_j))$_
.=where_link($s++,$W["col"],$W["val"],$W["op"]);}foreach($_j
as$He=>$W)$_
.=where_link($s++,$He,$W);}$Wd=select_value($X,$_,$l,$Yi);$u=bracket_escape($Aj);$t=h("val[$u][".bracket_escape($x)."]");$ih=idx(idx($_POST["val"],$u),bracket_escape($x));$Fj=idx($l["privileges"],"update");$tc=!is_array($K[$x])&&!is_blob($l)&&is_utf8($X)&&$L[$Mf][$x]==$X&&!$xd[$x]&&!$l["generated"]&&$Fj;$U=(preg_match('~^(AVG|MIN|MAX)\((.+)\)~',$c,$B)?$m[idf_unescape($B[2])]["type"]:$l["type"]);$Xi=preg_match('~text|json|lob~',$U);$_e=preg_match(number_type(),$U)||preg_match('~^(CHAR_LENGTH|ROUND|FLOOR|CEIL|TIME_TO_SEC|COUNT|SUM)\(~',$c);echo"<td id='$t'".($_e&&($X===null||is_numeric(strip_tags($Wd))||$U=="money")?" class='number'":"");if(($_GET["modify"]&&$tc&&$X!==null)||$ih!==null){$Fd=h($ih!==null?$ih:$X);echo">".($Xi?"<textarea name='$t' cols='30' rows='".(substr_count($X,"\n")+1)."'>$Fd</textarea>":"<input name='$t' value='$Fd' size='$Xe[$x]'>");}else{$ff=strpos($Wd,"<i>…</i>");echo($Fj?" data-text='".($ff?2:($Xi?1:0))."'".($tc?"":" data-warning='".'Use edit link to modify this value.'."'"):"").">$Wd";}}next($M);}if($Ea)echo"<td>";adminer()->backwardKeysPrint($Ea,$L[$Mf]);echo"</tr>\n";}if(is_ajax())exit;echo"</table>\n","</div>\n";}if(!is_ajax()){if($L||$F||$Kd){$Kc=true;if($_GET["page"]!="last"){if(!$z||(count($L)<$z&&($L||!$F)))$sd=($F?$F*$z:0)+count($L);elseif(JUSH!="sql"||!$ze){$sd=($ze?false:found_rows($S,$Z));if(intval($sd)<max(1e4,2*($F+1)*$z))$sd=first(slow_query(count_rows($a,$Z,$ze,$r)));elseif(JUSH=='sql'||JUSH=='pgsql')$Kc=false;}}if(!support("cursor"))$Kd=(($sd===false?count($L)+1:$sd-$F*$z)>$z);$Hg=($z&&($Kd||$F));if($Hg)echo($Kd?'<p><a href="'.h(pagination_href($F+1)).'" class="loadmore"'.on('click','selectLoadMore','Loading…').'>'.'Load more data'.'</a>':''),"\n";echo"<div class='footer'><div>\n";if($Hg){$pf=($sd===false?$F+($L?(count($L)>=$z?2:1):0):floor(($sd-1)/$z));echo"<fieldset><legend>".'Page'."</legend>";if(!support("cursor")){echo
pagination(0,$F).($F>5?" …":"");for($s=max(1,$F-4);$s<min($pf,$F+5);$s++)echo
pagination($s,$F);if($pf>0)echo($F+5<$pf?" …":""),($Kc&&$sd!==false?pagination($pf,$F):" <a href='".h(remove_from_uri("page")."&page=last")."' title='~$pf'>".'last'."</a>");}else
echo
pagination(0,$F).($F>1?" …":""),($F?pagination($F,$F):""),($Kd?pagination($F+1,$F)." …":"");echo"</fieldset>\n";}echo"<fieldset>","<legend>".'Whole result'."</legend>";$dc=($Kc?"":"~ ").$sd;$Le=($sd!==false?($Kc?"":"~ ").lang_format(array('%d row','%d rows'),$sd):"");echo
checkbox("all",1,0,$Le,on('click','countRows',$dc))."\n","</fieldset>\n";if(adminer()->selectCommandPrint())echo'<fieldset',($_GET["modify"]?'':" title='".'Ctrl+click on a value to modify it.'."'"),'>
<legend><a href=\'',h($_GET["modify"]?remove_from_uri("modify"):relative_uri()."&modify=1"),'\'>Modify</a></legend><div>
<input type=\'submit\' id=\'save\' value=\'Save\'',($_GET["modify"]?'':" class='jsonly' disabled"),'>
</div></fieldset>

<fieldset><legend>Selected <span id="selected"></span></legend><div>
<input type=\'submit\' name=\'edit\' value=\'Edit\'>
<input type=\'submit\' name=\'clone\' value=\'Clone\'>
<input type=\'submit\' name=\'delete\' value=\'Delete\'',confirm(),'>
</div></fieldset>
';$qd=adminer()->dumpFormat();foreach((array)$_GET["columns"]as$c){if($c["fun"]){unset($qd['sql']);break;}}if($qd){print_fieldset("export",'Export'." <span id='selected2'></span>");$Fg=adminer()->dumpOutput();echo($Fg?html_select("output",$Fg,$ja["output"])." ":""),html_select("format",$qd,$ja["format"])," <input type='submit' name='export' value='".'Export'."'>\n","</div></fieldset>\n";}adminer()->selectEmailPrint(array_filter($xc,'strlen'),$d);echo"</div></div>\n";}if(adminer()->selectImportPrint())echo"<p>","<a href='#import' class='toggle'>".'Import'."</a>","<span id='import'".($_POST["import"]?"":" class='hidden'").">: ",file_input(" name='csv_file'"," ".html_select("separator",array("csv"=>"CSV,","csv;"=>"CSV;","tsv"=>"TSV"),$ja["format"])." <input type='submit' name='import' value='".'Import'."'>"),"</span>";echo
input_token(),"</form>\n",(!$r&&$M?"":script("tableCheck();"));}}}if(is_ajax()){ob_end_clean();exit;}}elseif(isset($_GET["variables"])){$P=isset($_GET["status"]);page_header($P?'Status':'Variables');$Uj=($P?adminer()->showStatus():adminer()->showVariables());if(!$Uj)echo"<p class='message'>".'No rows.'."\n";else{echo"<table>\n";foreach($Uj
as$K){echo"<tr>";$x=array_shift($K);echo"<th><code class='jush-".JUSH.($P?"status":"set")."'>".h($x)."</code>";foreach($K
as$X)echo"<td>".nl_br(h($X));}echo"</table>\n";}}elseif(isset($_GET["script"])){header("Content-Type: application/json; charset=utf-8");if($_GET["script"]=="db"){$Gi=array("Data_length"=>0,"Index_length"=>0,"Data_free"=>0);foreach(table_status()as$D=>$S){json_row("Comment-$D",h($S["Comment"]));if(!is_view($S)||preg_match('~materialized~i',$S["Engine"])){foreach(array("Engine","Collation")as$x)json_row("$x-$D",h($S[$x]));foreach(array_keys($Gi+array("Auto_increment"=>0,"Rows"=>0))as$x){if(array_key_exists($x,$S))json_row("$x-$D",format_status($S,$x));if($S[$x]!=""&&isset($Gi[$x]))$Gi[$x]+=($S["Engine"]!="InnoDB"||$x!="Data_free"?$S[$x]:0);}}}if(function_exists('Adminer\db_status'))$Gi=db_status();foreach($Gi
as$x=>$X)json_row("sum-$x",format_number($X));json_row("");}elseif($_GET["script"]=="kill")connection()->query("KILL ".number($_POST["kill"]));else{foreach(count_tables(adminer()->databases(false))as$i=>$X){json_row("tables-$i",$X);json_row("size-$i",db_size($i));}json_row("");}exit;}else{$Ri=array_merge((array)$_POST["tables"],(array)$_POST["views"]);if($Ri&&!$k&&!$_POST["search"]){$I=true;$C="";if(JUSH=="sql"&&$_POST["tables"]&&count($_POST["tables"])>1&&($_POST["drop"]||$_POST["truncate"]||$_POST["copy"]))queries("SET foreign_key_checks = 0");if($_POST["truncate"]){if($_POST["tables"])$I=truncate_tables($_POST["tables"]);$C='Tables have been truncated.';}elseif($_POST["move"]){$I=move_tables((array)$_POST["tables"],(array)$_POST["views"],$_POST["target"]);$C='Tables have been moved.';}elseif($_POST["copy"]){$I=copy_tables((array)$_POST["tables"],(array)$_POST["views"],$_POST["target"]);$C='Tables have been copied.';}elseif($_POST["drop"]){if($_POST["views"])$I=drop_views($_POST["views"]);if($I&&$_POST["tables"])$I=drop_tables($_POST["tables"]);$C='Tables have been dropped.';}elseif(JUSH=="sqlite"&&$_POST["check"]){foreach((array)$_POST["tables"]as$R){foreach(get_rows("PRAGMA integrity_check(".q($R).")")as$K)$C
.="<b>".h($R)."</b>: ".h($K["integrity_check"])."<br>";}}elseif(JUSH!="sql"){$I=(JUSH=="sqlite"?queries("VACUUM"):apply_queries("VACUUM".($_POST["optimize"]?" ANALYZE":""),(array)$_POST["tables"]));$C='Tables have been optimized.';}elseif(!$_POST["tables"])$C='No tables.';elseif($I=queries(($_POST["optimize"]?"OPTIMIZE":($_POST["check"]?"CHECK":($_POST["repair"]?"REPAIR":"ANALYZE")))." TABLE ".implode(", ",array_map('Adminer\idf_escape',$_POST["tables"])))){while($K=$I->fetch_assoc())$C
.="<b>".h($K["Table"])."</b>: ".h($K["Msg_text"])."<br>";}queries_redirect($_SERVER["REQUEST_URI"],$C,$I);}page_header(($_GET["ns"]==""?'Database'.": ".h(DB):'Schema'.": ".h($_GET["ns"])),$k,true);if(adminer()->homepage()){if($_GET["ns"]!==""){$E=$_GET["order"];$ud=($E||support("fast_status"));echo"<div>\n","<h3 id='tables-views'>".'Tables and views'."</h3>\n";$Qi=($ud?table_status():tables_list());if(!$Qi)echo"<p class='message'>".'No tables.'."\n";else{echo"<form action='' method='post'>\n";if(support("table")){echo"<fieldset><legend>".'Search data in tables'." <span id='selected2'></span></legend><div>",html_select("op",adminer()->operators(),idx($_POST,"op",JUSH=="elastic"?"should":"LIKE %%"))," <input type='search' name='query' value='".h($_POST["query"])."'".on('keydown','submitKeydown','search').">"," <input type='submit' name='search' value='".'Search'."'>\n","</div></fieldset>\n";if(!$k&&$_POST["search"]&&$_POST["query"]!=""){$_GET["where"][0]["op"]=$_POST["op"];search_tables();}}echo"<div class='scrollable'>\n","<table class='nowrap checkable odds'".on('click','tableClick').on('dblclick','tableClick').">\n",'<thead><tr class="wrap">','<td class="hover"><input id="check-all" type="checkbox" class="jsonly" title="'.'All'.'"'.on('click','formCheck','^(tables|views)\[').'>','<th'.(!$E&&JUSH!='sqlite'?" aria-sort='ascending'":'').'><a href="'.h(substr(ME,0,-1)).'">'.'Table'.'</a>';$d=array("Engine"=>array('Engine'.doc_link(array('sql'=>'storage-engines.html'))));if(collations())$d["Collation"]=array('Collation'.doc_link(array('sql'=>'charset-charsets.html','mariadb'=>'supported-character-sets-and-collations/')));if(function_exists('Adminer\alter_table'))$d["Data_length"]=array('Data Length'.doc_link(array('sql'=>'show-table-status.html',)),"create",'Alter table',);if(support("indexes"))$d["Index_length"]=array('Index Length'.doc_link(array('sql'=>'show-table-status.html',)),"indexes",'Alter indexes',);$d["Data_free"]=array('Data Free'.doc_link(array('sql'=>'show-table-status.html')),"edit",'New item');if(function_exists('Adminer\alter_table'))$d["Auto_increment"]=array('Auto Increment'.doc_link(array('sql'=>'example-auto-increment.html','mariadb'=>'auto_increment/')),"auto_increment=1&create",'Alter table',);$d["Rows"]=array('Rows'.doc_link(array('sql'=>'show-table-status.html',)),"select",'Select data',);if(support("comment"))$d["Comment"]=array('Comment'.doc_link(array('sql'=>'show-table-status.html',)));$wa=array('Engine','Collation','Comment');foreach($d
as$x=>$c)echo"<th".($E==$x?" aria-sort='".(in_array($x,$wa)?"ascending":"descending")."'":"")."><a href='".h(ME)."order=$x'>$c[0]</a>";echo"<tbody>\n";if($E){uasort($Qi,function($fa,$Ba)use($E,$wa){$J=($fa[$E]<$Ba[$E]?-1:($fa[$E]>$Ba[$E]?1:0));return(in_array($E,$wa)?$J:-$J);});}$T=0;$Gi=array("Data_length"=>0,"Index_length"=>0,"Data_free"=>0);foreach($Qi
as$D=>$P){$Xj=($ud?is_view($P):$P!==null&&!preg_match('~table|sequence~i',$P));$P=($ud?$P:array('Engine'=>$P));$t=h("Table-".$D);echo'<tr><td class="hover">'.checkbox(($Xj?"views[]":"tables[]"),$D,in_array("$D",$Ri,true),"","","",$t),'<th>'.(support("table")||support("indexes")?"<a href='".h(ME)."table=".url_escape($D)."' title='".'Show structure'."' id='$t'>".h($D).'</a>':h($D));if($Xj&&!preg_match('~materialized~i',$P['Engine'])){$cj='View';echo'<td colspan="'.(count($d)-(support("comment")?2:1)).'">'.(support("view")?"<a href='".h(ME)."view=".url_escape($D)."' title='".'Alter view'."'>$cj</a>":$cj),"<td align='right'><a href='".h(ME)."select=".url_escape($D)."' title='".'Select data'."'>?</a>";if(support("comment"))echo'<td>'.h($P['Comment']);}else{if($ud){foreach(array_keys($Gi)as$x)$Gi[$x]+=($P["Engine"]!="InnoDB"||$x!="Data_free"?idx($P,$x):0);}foreach($d
as$x=>$c){$t=" id='$x-".h($D)."'";echo($c[1]?"<td align='right'><a href='".h(ME."$c[1]=").url_escape($D)."'$t title='$c[2]'>".format_status($P,$x)."</a>":"<td$t>".h(idx($P,$x,'?')));}$T++;}echo"\n";}echo"<tr><td class='hover'><th>".sprintf('%d in total',count($Qi)),"<td>".h(JUSH=="sql"?get_val("SELECT @@default_storage_engine"):""),(collations()?"<td>".h(db_collation(DB,collations())):'');if($ud&&function_exists('Adminer\db_status'))$Gi=db_status();foreach($Gi
as$x=>$Fi)echo($d[$x]?"<td align='right' id='sum-$x'>".($ud?format_number($Fi):""):"");echo"\n","</table>\n",($ud?'':script("ajaxSetHtml('".js_escape(ME)."script=db');")),"</div>\n";if(!information_schema(DB)){$Rj="<input type='submit' value='".'Vacuum'."'".on_help("VACUUM")."> ";$qg="<input type='submit' name='optimize' value='".'Optimize'."'".on_help(JUSH=="sql"?"OPTIMIZE TABLE":"VACUUM ANALYZE")."> ";$nh=(JUSH=="sqlite"?$Rj."<input type='submit' name='check' value='".'Check'."'".on_help("PRAGMA integrity_check")."> ":(JUSH=="pgsql"?$Rj.$qg:(JUSH=="sql"?"<input type='submit' value='".'Analyze'."'".on_help("ANALYZE TABLE")."> ".$qg."<input type='submit' name='check' value='".'Check'."'".on_help("CHECK TABLE")."> "."<input type='submit' name='repair' value='".'Repair'."'".on_help("REPAIR TABLE")."> ":""))).(function_exists('Adminer\truncate_tables')?"<input type='submit' name='truncate' value='".'Truncate'."'".confirm().on_help(JUSH=="sqlite"?"DELETE":"TRUNCATE".(JUSH=="pgsql"?"":" TABLE"))."> ":"").(function_exists('Adminer\drop_tables')?"<input type='submit' name='drop' value='".'Drop'."'".confirm().on_help("DROP TABLE").">":"");echo($nh?"<div class='footer'><div>\n<fieldset><legend>".'Selected'." <span id='selected'></span></legend><div>$nh\n</div></fieldset>\n":"");$h=(support("scheme")?adminer()->schemas():adminer()->databases());if(count($h)!=1&&function_exists('Adminer\move_tables')){echo"<fieldset><legend>".'Move to other database'." <span id='selected3'></span></legend><div>";$i=(isset($_POST["target"])?$_POST["target"]:(support("scheme")?$_GET["ns"]:DB));echo($h?html_select("target",$h,$i):'<input name="target" value="'.h($i).'" autocapitalize="off">'),"</label> <input type='submit' name='move' value='".'Move'."'>",(support("copy")?" <input type='submit' name='copy' value='".'Copy'."'> ".checkbox("overwrite",1,$_POST["overwrite"],'overwrite'):""),"</div></fieldset>\n";}echo"<input type='hidden' name='all' value=''".on('click','countTables',$T).">\n",input_token(),"</div></div>\n";}echo"</form>\n",script("tableCheck();");}echo(function_exists('Adminer\alter_table')?"<p class='links hover'><a href='".h(ME)."create='>".'Create table'."</a>\n":''),(support("view")?"<a href='".h(ME)."view='>".'Create view'."</a>\n":""),"</div>\n";if(support("routine")){echo"<div>\n","<h3 id='routines'>".'Routines'."</h3>\n";$Sh=routines();if($Sh){echo"<table class='odds'>\n",'<thead><tr><th>'.'Name'.'<td>'.'Type'.'<td>'.'Return type'."<td class='hover'><tbody>\n";foreach($Sh
as$K){$D=($K["SPECIFIC_NAME"]==$K["ROUTINE_NAME"]?"":"&name=".url_escape($K["ROUTINE_NAME"]));echo'<tr>','<th><a href="'.h(ME.($K["ROUTINE_TYPE"]!="PROCEDURE"?'callf=':'call=').url_escape($K["SPECIFIC_NAME"]).$D).'">'.h($K["ROUTINE_NAME"]).'</a>','<td>'.h($K["ROUTINE_TYPE"]),'<td>'.h($K["DTD_IDENTIFIER"]),'<td class="hover"><a href="'.h(ME.($K["ROUTINE_TYPE"]!="PROCEDURE"?'function=':'procedure=').url_escape($K["SPECIFIC_NAME"]).$D).'">'.'Alter'."</a>";}echo"</table>\n";}echo'<p class="links hover">'.(support("procedure")?'<a href="'.h(ME).'procedure=">'.'Create procedure'.'</a>':'').'<a href="'.h(ME).'function=">'.'Create function'."</a>\n","</div>\n";}if(support("event")){echo"<div>\n","<h3 id='events'>".'Events'."</h3>\n";$L=get_rows("SHOW EVENTS");if($L){echo"<table>\n","<thead><tr><th>".'Name'."<td>".'Schedule'."<td>".'Start'."<td>".'End'."<td><tbody>\n";foreach($L
as$K)echo"<tr>","<th>".h($K["Name"]),"<td>".($K["Execute at"]?'At given time'."<td>".h($K["Execute at"]):'Every'." ".h($K["Interval value"])." ".h($K["Interval field"])."<td>".h($K["Starts"])),"<td>".h($K["Ends"]),'<td><a href="'.h(ME).'event='.url_escape($K["Name"]).'">'.'Alter'.'</a>';echo"</table>\n";$Ic=get_val("SELECT @@event_scheduler");if($Ic&&$Ic!="ON")echo"<p class='error'><code class='jush-sqlset'>event_scheduler</code>: ".h($Ic)."\n";}echo'<p class="links hover"><a href="'.h(ME).'event=">'.'Create event'."</a>\n","</div>\n";}}}}page_footer();