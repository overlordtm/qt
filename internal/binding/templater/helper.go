package templater

import (
	"sort"
	"strings"

	"github.com/therecipe/qt/internal/binding/parser"
	"github.com/therecipe/qt/internal/utils"
)

func functionIsSupported(_ *parser.Class, f *parser.Function) bool {

	switch {
	case
		(f.ClassName() == "QAccessibleObject" || f.ClassName() == "QAccessibleInterface" || f.ClassName() == "QAccessibleWidget" || //QAccessible::State -> quint64
			f.ClassName() == "QAccessibleStateChangeEvent") && (f.Name == "state" || f.Name == "changedStates" || f.Name == "m_changedStates" || f.Name == "setM_changedStates" || f.Meta == parser.CONSTRUCTOR),

		f.Fullname == "QPixmapCache::find" && f.OverloadNumber == "4", //Qt::Key -> int
		(f.Fullname == "QPixmapCache::remove" || f.Fullname == "QPixmapCache::insert") && f.OverloadNumber == "2",
		f.Fullname == "QPixmapCache::replace",

		f.Fullname == "QNdefFilter::appendRecord" && !f.Overload, //QNdefRecord::TypeNameFormat -> uint

		f.ClassName() == "QSimpleXmlNodeModel" && f.Meta == parser.CONSTRUCTOR,

		f.Fullname == "QSGMaterialShader::attributeNames",

		f.ClassName() == "QVariant" && (f.Name == "value" || f.Name == "canConvert"), //needs template

		f.Fullname == "QNdefRecord::isRecordType", f.Fullname == "QScriptEngine::scriptValueFromQMetaObject", //needs template
		f.Fullname == "QScriptEngine::fromScriptValue", f.Fullname == "QJSEngine::fromScriptValue",

		f.ClassName() == "QMetaType" && //needs template
			(f.Name == "hasRegisteredComparators" || f.Name == "registerComparators" ||
				f.Name == "hasRegisteredConverterFunction" || f.Name == "registerConverter" ||
				f.Name == "registerEqualsComparator"),

		parser.CurrentState.ClassMap[f.ClassName()].Module == parser.MOC && f.Name == "metaObject", //needed for qtmoc

		f.Fullname == "QSignalBlocker::QSignalBlocker" && f.OverloadNumber == "3", //undefined symbol

		(parser.CurrentState.ClassMap[f.ClassName()].IsSubClassOf("QCoreApplication") ||
			f.ClassName() == "QAudioInput" || f.ClassName() == "QAudioOutput") && f.Name == "notify", //redeclared (name collision with QObject)

		f.Fullname == "QGraphicsItem::isBlockedByModalPanel", //** problem

		f.Name == "surfaceHandle", //QQuickWindow && QQuickView //unsupported_cppType(QPlatformSurface)

		f.Name == "readData", f.Name == "QNetworkReply", //TODO: char*

		f.Name == "QDesignerFormWindowInterface" || f.Name == "QDesignerFormWindowManagerInterface" || f.Name == "QDesignerWidgetBoxInterface", //unimplemented virtual

		f.Fullname == "QNdefNfcSmartPosterRecord::titleRecords", //T<T> output with unsupported output for *_atList
		f.Fullname == "QHelpEngineCore::filterAttributeSets", f.Fullname == "QHelpSearchEngine::query", f.Fullname == "QHelpSearchQueryWidget::query",
		f.Fullname == "QPluginLoader::staticPlugins", f.Fullname == "QSslConfiguration::ellipticCurves", f.Fullname == "QSslConfiguration::supportedEllipticCurves",
		f.Fullname == "QTextFormat::lengthVectorProperty", f.Fullname == "QTextTableFormat::columnWidthConstraints",

		strings.Contains(f.Access, "unsupported"):
		{
			if !strings.Contains(f.Access, "unsupported") {
				f.Access = "unsupported_isBlockedFunction"
			}
			return false
		}
	}

	if strings.ContainsAny(f.Signature, "<>") {
		if parser.IsPackedList(f.Output) {
			for _, p := range f.Parameters {
				if strings.ContainsAny(p.Value, "<>") {
					if !strings.Contains(f.Access, "unsupported") {
						f.Access = "unsupported_isBlockedFunction"
					}
					return false
				}
			}
		} else {
			if !strings.Contains(f.Access, "unsupported") {
				f.Access = "unsupported_isBlockedFunction"
			}
			return false
		}
	}

	if f.Default {
		return functionIsSupportedDefault(f)
	}

	if parser.CurrentState.Minimal {
		return f.Export || f.Meta == parser.DESTRUCTOR || f.Fullname == "QObject::destroyed" || strings.HasPrefix(f.Name, parser.TILDE)
	}

	return true
}

func functionIsSupportedDefault(f *parser.Function) bool {
	switch f.Fullname {
	case
		"QAnimationGroup::updateCurrentTime", "QAnimationGroup::duration",
		"QAbstractProxyModel::columnCount", "QAbstractTableModel::columnCount",
		"QAbstractListModel::data", "QAbstractTableModel::data",
		"QAbstractProxyModel::index", "QAbstractProxyModel::parent",
		"QAbstractListModel::rowCount", "QAbstractProxyModel::rowCount", "QAbstractTableModel::rowCount",

		"QPagedPaintDevice::paintEngine", "QAccessibleObject::childCount",
		"QAccessibleObject::indexOfChild", "QAccessibleObject::role",
		"QAccessibleObject::text", "QAccessibleObject::child",
		"QAccessibleObject::parent",

		"QAbstractGraphicsShapeItem::paint", "QGraphicsObject::paint",
		"QLayout::sizeHint", "QAbstractGraphicsShapeItem::boundingRect",
		"QGraphicsObject::boundingRect", "QGraphicsLayout::sizeHint",

		"QSimpleXmlNodeModel::typedValue", "QSimpleXmlNodeModel::documentUri",
		"QSimpleXmlNodeModel::compareOrder", "QSimpleXmlNodeModel::nextFromSimpleAxis",
		"QSimpleXmlNodeModel::kind", "QSimpleXmlNodeModel::root",

		"QAbstractPlanarVideoBuffer::unmap", "QAbstractPlanarVideoBuffer::mapMode",

		"QSGDynamicTexture::bind", "QSGDynamicTexture::hasMipmaps",
		"QSGDynamicTexture::textureSize", "QSGDynamicTexture::hasAlphaChannel",
		"QSGDynamicTexture::textureId",

		"QModbusClient::open", "QModbusClient::close", "QModbusServer::open", "QModbusServer::close",

		"QSimpleXmlNodeModel::name":

		{
			return false
		}
	}

	//needed for moc
	if parser.CurrentState.ClassMap[f.ClassName()].IsSubClassOf("QCoreApplication") && (f.Name == "autoMaximizeThreshold" || f.Name == "setAutoMaximizeThreshold") {
		return false
	}

	if parser.CurrentState.Minimal {
		return f.Export
	}

	return true
}

func classIsSupported(c *parser.Class) bool {
	if c == nil {
		return false
	}

	switch c.Name {
	case
		"QString", "QStringList", //mapped to primitive

		"QExplicitlySharedDataPointer", "QFuture", "QDBusPendingReply", "QDBusReply", "QFutureSynchronizer", //needs template
		"QGlobalStatic", "QMultiHash", "QQueue", "QMultiMap", "QScopedPointer", "QSharedDataPointer",
		"QScopedArrayPointer", "QSharedPointer", "QThreadStorage", "QScopedValueRollback", "QVarLengthArray",
		"QWeakPointer", "QWinEventNotifier",

		"QFlags", "QException", "QStandardItemEditorCreator", "QSGSimpleMaterialShader", "QGeoCodeReply", "QFutureWatcher", //other
		"QItemEditorCreator", "QGeoCodingManager", "QGeoCodingManagerEngine", "QQmlListProperty",

		"QPlatformGraphicsBuffer", "QPlatformSystemTrayIcon", "QRasterPaintEngine", "QSupportedWritingSystems", "QGeoLocation", //file not found or QPA API
		"QAbstractOpenGLFunctions",

		"QProcess", "QProcessEnvironment": //TODO: iOS

		{
			c.Access = "unsupported_isBlockedClass"
			return false
		}
	}

	switch {
	case
		strings.HasPrefix(c.Name, "QOpenGL"), strings.HasPrefix(c.Name, "QPlace"), //file not found or QPA API

		strings.HasPrefix(c.Name, "QAtomic"), //other

		strings.HasSuffix(c.Name, "terator"), strings.Contains(c.Brief, "emplate"): //needs template

		{
			c.Access = "unsupported_isBlockedClass"
			return false
		}
	}

	if parser.CurrentState.Minimal {
		return c.Export
	}

	return true
}

func hasUnimplementedPureVirtualFunctions(className string) bool {
	for _, f := range parser.CurrentState.ClassMap[className].Functions {
		var f = *f
		cppFunction(&f)

		if f.Virtual == parser.PURE && !functionIsSupported(parser.CurrentState.ClassMap[className], &f) {
			return true
		}
	}
	return false
}

func ShouldBuild(module string) bool {
	return true //Build[module]
}

var Build = map[string]bool{
	"Core":          false,
	"AndroidExtras": false,
	"Gui":           false,
	"Network":       false,
	"Xml":           false,
	"DBus":          false,
	"Nfc":           false,
	"Script":        false,
	"Sensors":       false,
	"Positioning":   false,
	"Widgets":       false,
	"Sql":           false,
	"MacExtras":     false,
	"Qml":           false,
	"WebSockets":    false,
	"XmlPatterns":   false,
	"Bluetooth":     false,
	"WebChannel":    false,
	"Svg":           false,
	"Multimedia":    false,
	"Quick":         false,
	"Help":          false,
	"Location":      false,
	"ScriptTools":   false,
	"UiTools":       false,
	"X11Extras":     false,
	"WinExtras":     false,
	"WebEngine":     false,
	"TestLib":       false,
	"SerialPort":    false,
	"SerialBus":     false,
	"PrintSupport":  false,
	//"PlatformHeaders": false,
	"Designer": false,
	"Scxml":    false,
	"Gamepad":  false,

	"Purchasing": false,
	//"DataVisualization": false,
	//"Charts":            false,
	//"Quick2DRenderer":   false,

	"Sailfish": false,
}

func isGeneric(f *parser.Function) bool {

	if f.ClassName() == "QAndroidJniObject" {
		switch f.Name {
		case
			"callMethod",
			"callStaticMethod",

			"getField",
			//"setField", -> uses interface{} if not generic

			"getStaticField",
			//"setStaticField", -> uses interface{} if not generic

			"getObjectField",

			"getStaticObjectField",

			"callObjectMethod",
			"callStaticObjectMethod":
			{
				return true
			}

		case "setStaticField":
			{
				if f.OverloadNumber == "2" || f.OverloadNumber == "4" {
					return true
				}
			}
		}
	}

	return false
}

func classNeedsCallbackFunctions(class *parser.Class) bool {

	for _, function := range class.Functions {
		if function.Virtual == parser.IMPURE || function.Virtual == parser.PURE || function.Meta == parser.SIGNAL || function.Meta == parser.SLOT {
			return true
		}
	}

	return false
}

func shortModule(module string) string {
	return strings.ToLower(strings.TrimPrefix(module, "Qt"))
}

func getSortedClassNamesForModule(module string) []string {
	var output = make([]string, 0)
	for _, class := range parser.CurrentState.ClassMap {
		if class.Module == module {
			output = append(output, class.Name)
		}
	}
	sort.Stable(sort.StringSlice(output))
	return output
}

func getSortedClassesForModule(module string) []*parser.Class {
	var (
		classNames = getSortedClassNamesForModule(module)
		output     = make([]*parser.Class, len(classNames))
	)
	for i, name := range classNames {
		output[i] = parser.CurrentState.ClassMap[name]
	}
	return output
}

func classNeedsDestructor(c *parser.Class) bool {
	for _, f := range c.Functions {
		if f.Meta == parser.DESTRUCTOR {
			return false
		}
	}
	return true
}

func UseStub() bool {
	return utils.QT_STUB() && !parser.CurrentState.Minimal && !parser.CurrentState.Moc && !(parser.CurrentState.CurrentModule == "AndroidExtras" || parser.CurrentState.CurrentModule == "Sailfish")
}
