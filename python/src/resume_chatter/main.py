from langchain_openai import ChatOpenAI

import logging

logger = logging.getLogger(__name__)


if __name__=="__main__":
    model = "gpt-4o-mini"
    logger.warning(f"using model: {model}")
    llm = ChatOpenAI(model=model)
    res = llm.invoke("how can langsmith help with testing?")
    print(res)
