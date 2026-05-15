<template>
  <div class="container">
    <div class="intro-overlay">
      <div class="intro-beam" />
      <div class="intro-content">
        <div class="intro-kicker">观沧海</div>
        <div class="intro-title">观潮</div>
        <div class="intro-subtitle">弄潮头</div>
        <div class="intro-desc">多路并发，立体研判，直达竞品决策现场</div>
      </div>
      <div class="cta-wrap">
        <button type="button" class="launch-btn" @click="handleEnter">
          进入竞品分析
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useUserStore } from '@/store';
  import { setToken } from '@/utils/auth';

  const userStore = useUserStore();

  const handleEnter = async () => {
    // 模拟默认账号登录态，满足路由守卫对 token 的检查
    setToken('12345');
    localStorage.setItem('userRole', 'admin');
    userStore.setInfo({
      name: '分析师',
      role: 'admin',
    });
    window.location.assign('/analysis/new');
  };
</script>

<style lang="less" scoped>
  .container {
    position: relative;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    height: 100vh;
    background: linear-gradient(
      180deg,
      rgba(2, 10, 26, 1) 0%,
      rgba(3, 21, 47, 0.99) 48%,
      rgba(5, 48, 96, 1) 100%
    );
  }

  .intro-overlay {
    position: absolute;
    inset: 0;
    z-index: 10;
    overflow: hidden;
    background:
      radial-gradient(
        circle at 50% 48%,
        rgba(136, 221, 255, 0.28) 0%,
        rgba(136, 221, 255, 0.17) 10%,
        rgba(136, 221, 255, 0.07) 18%,
        rgba(136, 221, 255, 0.02) 26%,
        rgba(136, 221, 255, 0) 36%
      ),
      linear-gradient(
        180deg,
        rgba(2, 10, 26, 1) 0%,
        rgba(3, 21, 47, 0.99) 48%,
        rgba(5, 48, 96, 1) 100%
      );
    animation: center-glow 1.9s ease-out forwards;
  }

  .intro-beam {
    position: absolute;
    top: -20%;
    left: 50%;
    width: 1px;
    height: 62%;
    background: linear-gradient(180deg, rgba(255, 255, 255, 0), rgba(173, 231, 255, 0.85), rgba(255, 255, 255, 0));
    opacity: 0;
    transform: translateX(-50%);
    box-shadow: 0 0 20px rgba(130, 220, 255, 0.4);
    animation: beam-drop 1.2s ease-out 0.18s forwards;
  }

  .intro-content {
    position: absolute;
    top: 42%;
    left: 50%;
    z-index: 2;
    transform: translate(-50%, -50%);
    text-align: center;
    color: #fff;
    animation: title-rise 1.9s ease-out forwards;
  }

  .intro-kicker,
  .intro-subtitle {
    font-size: 18px;
    letter-spacing: 14px;
    color: rgba(235, 245, 255, 0.78);
  }

  .intro-title {
    margin: 16px 0 18px;
    font-size: 96px;
    font-weight: 700;
    letter-spacing: 24px;
    text-indent: 24px;
    text-shadow:
      0 0 30px rgba(77, 171, 247, 0.4),
      0 0 60px rgba(77, 171, 247, 0.2);
  }

  .intro-desc {
    margin-top: 22px;
    font-size: 14px;
    letter-spacing: 4px;
    color: rgba(222, 238, 255, 0.56);
  }

  .cta-wrap {
    position: absolute;
    left: 50%;
    bottom: 12%;
    z-index: 2;
    display: flex;
    align-items: center;
    transform: translateX(-50%);
  }

  .launch-btn {
    min-width: 240px;
    height: 56px;
    padding: 0 36px;
    border: 1px solid rgba(179, 234, 255, 0.42);
    border-radius: 999px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.04)),
      linear-gradient(90deg, rgba(49, 138, 214, 0.38), rgba(93, 192, 255, 0.26));
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    color: #f4fbff;
    font-size: 16px;
    font-weight: 600;
    letter-spacing: 3px;
    cursor: pointer;
    box-shadow:
      0 20px 44px rgba(3, 14, 34, 0.28),
      0 0 0 1px rgba(169, 230, 255, 0.12) inset,
      0 0 22px rgba(118, 208, 255, 0.28),
      0 0 48px rgba(118, 208, 255, 0.18);
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease,
      filter 0.2s ease,
      background 0.2s ease,
      border-color 0.2s ease;
  }

  .launch-btn:hover {
    transform: translateY(-1px);
    border-color: rgba(196, 240, 255, 0.58);
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.24), rgba(255, 255, 255, 0.06)),
      linear-gradient(90deg, rgba(56, 154, 231, 0.44), rgba(110, 205, 255, 0.3));
    box-shadow:
      0 24px 48px rgba(3, 14, 34, 0.34),
      0 0 0 1px rgba(196, 240, 255, 0.16) inset,
      0 0 30px rgba(118, 208, 255, 0.34),
      0 0 62px rgba(118, 208, 255, 0.24);
    filter: brightness(1.06);
  }

  .launch-btn:active {
    transform: translateY(1px);
  }

  @keyframes title-rise {
    0% {
      opacity: 0;
      transform: translate(-50%, -44%) scale(0.96);
      filter: blur(14px);
    }

    44% {
      opacity: 1;
      transform: translate(-50%, -52%) scale(1.03);
      filter: blur(0);
    }

    100% {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
      filter: blur(0);
    }
  }

  @keyframes center-glow {
    0% {
      opacity: 0.82;
      filter: saturate(0.92);
    }

    100% {
      opacity: 1;
      filter: saturate(1);
    }
  }

  @keyframes beam-drop {
    0% {
      opacity: 0;
      transform: translateX(-50%) translateY(-24px);
    }

    40% {
      opacity: 1;
    }

    100% {
      opacity: 0.15;
      transform: translateX(-50%) translateY(0);
    }
  }
</style>
