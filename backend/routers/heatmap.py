from fastapi import APIRouter

router = APIRouter()

@router.post("/generate_heatmap", summary="生成Bilibili观看历史热力图")
def api_generate_heatmap():
    from scripts.heatmap_visualizer import generate_heatmap
    return generate_heatmap()
