import io
from typing import List

import clip
import torch
from PIL import Image
from fastapi import FastAPI, File

app = FastAPI()
device = "cuda" if torch.cuda.is_available() else "cpu"
model, preprocess = clip.load("ViT-B/32", device=device)


@app.get("/encode_text/{text}", response_model=List[float])
def encode_text(text: str) -> List[float]:
    text_inputs = clip.tokenize(text).to(device)
    encoded_text = model.encode_text(text_inputs)
    return encoded_text[0].cpu().tolist()


@app.get("/encode_image", response_model=List[float])
def encode_image(image: bytes = File(...)) -> List[float]:
    image = preprocess(Image.open(io.BytesIO(image))).unsqueeze(0).to(device)
    encoded_image = model.encode_image(image)
    return encoded_image[0].cpu().tolist()


if __name__ == "__main__":
    encode_text('photoprism is great')
