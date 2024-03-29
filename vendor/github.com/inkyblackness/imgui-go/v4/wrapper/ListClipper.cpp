#include "ConfiguredImGui.h"

#include "ListClipper.h"
#include "WrapperConverter.h"

static void importValue(ImGuiListClipper &out, IggListClipper const &in);
static void exportValue(IggListClipper &out, ImGuiListClipper const &in);

typedef TypeWrapper<ImGuiListClipper, IggListClipper> ListClipperWrapper;

IggBool iggListClipperStep(IggListClipper *clipper)
{
   ImGuiListClipper imguiClipper;
   importValue(imguiClipper, *clipper);
   IggBool returnValue = imguiClipper.Step() ? 1 : 0;
   exportValue(*clipper, imguiClipper);
   // needs to be done to prevent assert fail, we don't call end because the cursor will move.
   imguiClipper.ItemsCount = -1;
   return returnValue;
}

void iggListClipperBegin(IggListClipper *clipper, int items_count, float items_height)
{
   ImGuiListClipper imguiClipper;
   imguiClipper.Begin(items_count, items_height);
   exportValue(*clipper, imguiClipper);
   // needs to be done to prevent assert fail, we don't call end because the cursor will move.
   imguiClipper.ItemsCount = -1;
}

void iggListClipperEnd(IggListClipper *clipper)
{
   ImGuiListClipper imguiClipper;
   importValue(imguiClipper, *clipper);
   imguiClipper.End();
   exportValue(*clipper, imguiClipper);
}

static void importValue(ImGuiListClipper &out, IggListClipper const &in)
{
   out.DisplayStart = in.DisplayStart;
   out.DisplayEnd = in.DisplayEnd;
   out.ItemsCount = in.ItemsCount;

   out.ItemsHeight = in.ItemsHeight;
   out.StartPosY = in.StartPosY;
   out.TempData = in.TempData;
}

static void exportValue(IggListClipper &out, ImGuiListClipper const &in)
{
   out.DisplayStart = in.DisplayStart;
   out.DisplayEnd = in.DisplayEnd;
   out.ItemsCount = in.ItemsCount;

   out.ItemsHeight = in.ItemsHeight;
   out.StartPosY = in.StartPosY;
   out.TempData = in.TempData;
}
