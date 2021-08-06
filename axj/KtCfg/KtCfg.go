package KtCfg

import (
	"axj/Kt"
	"axj/KtStr"
	"os"
	"strings"
)

// 配置字典别名
type Cfg map[string]interface{}

// 配置获取
func Get(cfg Cfg, name string) interface{} {
	val := Kt.If(cfg == nil, nil, cfg[name])
	if val == nil {
		if len(name) > 1 {
			chr := name[0]
			switch chr {
			case '%':
				return os.Getenv(name[1:])
				break
			case '@':
			case '$':
				return Kt.If(cfg == nil, nil, cfg[name[1:]])
				break
			}
		}
	}

	return val
}

func  GetExp(exp string, strict bool, cfg Cfg) string {
return GetExpRemain(exp, strict, false, cfg);
}

func GetExpRemain(exp string, strict bool,  remain bool, cfg Cfg) string {
 index := strings.Index(exp, "${");
length := len(exp)
	if (index >= 0 && index < length - 2) {

		var sb strings.Builder
 sIndex := 0;
		for {
if (index > sIndex) {
sb.WriteString(exp[sIndex: index]);

} else if (index < sIndex) {
if (index < 0) {
if (length > sIndex) {
	sb.WriteString(exp[sIndex: length]);
}
}

break;
}

			sIndex = KtStr.Index(exp, "${", index)
if (sIndex < 0) {
	sb.WriteString(exp[index: length]);
break;
}

index += 2;
if (index < sIndex) {
 val := exp[index: sIndex];
var valD interface{} = nil;
	var value interface{} = nil;
if (strings.Index(val, "?") > 0) {
// 支持二运运算|三元运算
List<String> vals = KtStr.splitStr(val, "?:", true, 0);
val = vals.get(0);
if (vals.size() == 3) {
strict = false;
valD = getValClass(val, cfgMap, boolean.class, false, null) ? vals.get(1): vals.get(2);

} else {
value = getVal(val, cfgMap);
valD = vals.get(1);
}

} else {
value = getVal(val, cfgMap);
}

if (value == null) {
if (valD != null) {
sb.append(valD);

} else if (strict) {
return null;

} else if (remain) {
sb.append("${");
sb.append(val);
sb.append("}");
}

} else {
sb.append(value);
}
}

index = exp.indexOf("${", sIndex++);
}

exp = sb.toString();
if (exp.indexOf("$$") >= 0) {
exp = exp.replace("$$", "$");
}
}

return exp;
}

public static Object getValType(String name, Map<String, Object> cfgMap, Type toType, Object dVal, String tName) {
Object val = getVal(name, cfgMap);
if (val == null) {
return KtCvt.nullTo(toType, dVal, null);
}

Class toCls = KtCls.raw(toType);
if (!KtCvt.is(val, toCls)) {
if (toCls.isAssignableFrom(ArrayList.class)) {
List<Object> list = new ArrayList<Object>();
list.add(val);
val = list;

} else {
val = SBinder.ME.toCvt(val, toType, dVal, tName, null);
}

if (val != null && cfgMap.containsKey(name)) {
// CLoader转换不保存;避免内存泄漏
if (!(val.getClass().getClassLoader() instanceof CLoader)) {
cfgMap.put(name, val);
}
}
}

return val;
}

public static <T> T getValClass(String name, Map<String, Object> cfgMap, Class<T> toType, T dVal, String tName) {
return (T) getValType(name, cfgMap, toType, dVal, tName);
}

static final char[] SPLITS = new char[]{'=', ':'}

static KtB.Func1<Void, String> readFunc(final Map<String, Object> cfgMap,
final Map<String, KtB.Func1<Void, String>> funcMap) {
return new KtB.Func1<Void, String>() {

private StringBuilder bBuilder;

private int bAppend;

private int yB;

private LinkedHashMap yMap;

private Stack<LinkedHashMap> yMaps;

@Override
public Void do1(String s) {
int length = s.length();
if (length < 1) {
return null;
}

char chr = s.charAt(0);
if (bBuilder == null) {
if (chr == '#' || chr == ';') {
return null;

} else if (chr == '{' && length == 2 && s.charAt(1) == '"') {
bBuilder = new StringBuilder();
bAppend = 1;
return null;
}

} else if (bAppend > 0) {
if (chr == '"' && length == 2 && s.charAt(1) == '}') {
bAppend = 0;

} else {
if (bAppend > 1) {
bBuilder.append("\r\n");

} else {
bAppend = 2;
}

bBuilder.append(getExp(s, false, cfgMap));
}

return null;
}

if (length < 3) {
return null;
}

int index = KtStr.indexAny(s, SPLITS);
if (index > 0 && index < length) {
String name;
chr = s.charAt(index - 1);
if (chr == '.' || chr == '#' || chr == ',' || chr == '+' || chr == '-') {
if (index < 1) {
return null;
}

name = s.substring(0, index - 1);

} else {
chr = 0;
name = s.substring(0, index);
}

name = name.trim();
length = name.length();
if (length == 0) {
return null;
}

// yml支持
if (s.charAt(index) == ':' && s.substring(index).trim().length() == 1) {
int b = index - length;
if (yB < b) {
if (yMap != null) {
if (yMaps == null) {
yMaps = new Stack<LinkedHashMap>();
}

yMaps.push(yMap);
}

LinkedHashMap map = new LinkedHashMap();
(yMap == null ? cfgMap: yMap).put(name, map);
yMap = map;

} else {
if (b == 0) {
// 根配置
if (yMaps != null) {
yMaps.clear();
}

yMap = new LinkedHashMap();
cfgMap.put(name, yMap);

} else {
if (yMaps == null || yMaps.empty()) {
yMap = null;

} else {
yMap = yMaps.pop();
}
}
}

yB = b;
return null;
}

int eIndex = index;
index = name.indexOf('|');
if (index > 0) {
if (length <= 1) {
return null;
}

String[] conds = KtStr.splitChr(name, '|', index + 1, 0);
name = name.substring(0, index).trim();
for (String cond : conds) {
index = cond.indexOf('&');
if (index > 0) {
String val = KtCvt.to(getVal(cond.substring(0, index), cfgMap), String.class, null);
if (val != null && KtStr.match(cond.substring(index + 1), false, val)) {
conds = null;
break;
}

} else if (getValClass(cond, cfgMap, boolean.class, false, null)) {
conds = null;
break;
}
}

if (conds != null) {
return null;
}
}

s = s.substring(eIndex + 1);
s = getExp(KtStr.toArg(s), false, cfgMap);
if (bBuilder != null) {
if (s.length() > 0) {
bBuilder.append("\r\n");
bBuilder.append(s);
}

s = bBuilder.toString();
bBuilder = null;
bAppend = 0;
}

if (funcMap != null && name.charAt(0) == '@') {
KtB.Func1<Void, String> func1 = funcMap.get(name);
if (func1 != null) {
func1.do1(s);
return null;
}
}

Map map = yMap == null ? cfgMap: yMap;
Object o;
switch (chr) {
case '.':
o = map.get(name);
map.put(name, o == null ? s: (o + s));
break;
case '#':
o = map.get(name);
map.put(name, o == null ? s: (o + "\r\n" + s));
break;
case ',':
o = getValClass(name, map, List.class, null, null);
if (o == null) {
o = new ArrayList<Object>();
map.put(name, o);
}
((List<Object>) o).addAll(KtStr.splitStr(s, ",;", true, 0));
break;
case '+':
o = getValClass(name, map, List.class, null, null);
if (o == null) {
o = new ArrayList<Object>();
map.put(name, o);
}
((List<Object>) o).add(s);
break;
case '-':
map.remove(name);
break;
default:
map.put(name, s);
break;
}
}

return null;
}
};
}

public static Map<String, Object> readStream(InputStream in, Map<String, Object> cfgMap,
Map<String, KtB.Func1<Void, String>> funcMap) {
if (in == null) {
return cfgMap;
}

try {
if (cfgMap == null) {
cfgMap = new HashMap<String, Object>();
}

KtIo.streamLine(in, null, readFunc(cfgMap, funcMap));

} catch (IOException e) {
KtA.throwEx(e);
}

return cfgMap;
}
