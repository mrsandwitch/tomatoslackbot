<template>
  <div class="clock-container">
    <div class="button-row">
      <button
        class="uk-button uk-button-default"
        v-on:click="clockStart('10m', true)"
      >
        10 (work)
      </button>
      <button
        class="uk-button uk-button-primary"
        v-on:click="clockStart('10m', false)"
      >
        10 (spare)
      </button>
    </div>
    <div class="button-row">
      <button
        class="uk-button uk-button-default"
        v-on:click="clockStart('25m', true)"
      >
        25 (work)
      </button>
      <button
        class="uk-button uk-button-primary"
        v-on:click="clockStart('25m', false)"
      >
        25 (spare)
      </button>
    </div>
  </div>

  <hr class="uk-divider-icon" />

  <div class="ongoing-clock-container">
    <div class="ongoing-clock-row">
      <div uk-countdown="date: 2021-10-10T12:27:56+00:00">
        <span class="uk-countdown-number uk-countdown-minutes"></span>
        :
        <span class="uk-countdown-number uk-countdown-seconds"></span>
      </div>
      <button class="uk-button uk-button-primary" name="stop" value="stop">
        stop
      </button>
    </div>
    <div class="ongoing-clock-row">
      <div uk-countdown="date: 2021-10-10T12:27:56+00:00">
        <span class="uk-countdown-number uk-countdown-minutes"></span>
        :
        <span class="uk-countdown-number uk-countdown-seconds"></span>
      </div>
      <button class="uk-button uk-button-primary" name="stop" value="stop">
        stop
      </button>
    </div>
  </div>

  <div class="uk-child-width-1-3@m uk-child-width-1-2@s records-container">
    <div v-for="item in items" v-bind:key="item" class="record-card-container">
      <div class="uk-card uk-card-default uk-card-body">
        <h3 class="uk-card-title">
          <div class="uk-card-badge uk-label">{{ item.Count }}</div>
          {{ item.Title }}
        </h3>
        <table class="uk-table uk-table-small uk-text-nowrap">
          <tbody>
            <tr v-for="readable in item.Readables" v-bind:key="readable">
              <td class="uk-width-1-4">{{ readable.Start }}</td>
              <td class="uk-width-1-4">{{ readable.Duration }}</td>
              <td v-bind:class="'uk-text-' + readable.Label">
                {{ readable.Tag }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import { ref, onMounted } from "vue";
const apiUrl = "http://localhost:8000";

export default {
  //   name: "Clock",
  //   props: {
  //     msg: String,
  //   },
  //data() {
  //  //return {
  //  //  items: null,
  //  //};
  //},
  setup() {
    const items = ref([]);

    const getRecords = async () => {
      try {
        items.value = await axios
          .get(apiUrl + "/api/records")
          .then((resp) => resp.data.Items);
      } catch (e) {
        console.log(e);
      }
    };

    onMounted(() => {
      getRecords();
    });

    return {
      items,
    };
  },
  methods: {
    async clockStart(time, forWork) {
      console.log((forWork ? "w" : "s") + " " + time);
      try {
        await axios.post(apiUrl + "/tomato", {
          ctlStr: (forWork ? "w" : "s") + " " + time,
        });
      } catch (e) {
        console.log(e);
      }
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.clock-container {
  padding: 40px 20px 0px 20px;
  display: flex;
  flex-direction: column;
}

.button-row {
  display: flex;
  justify-content: center;
  padding: 10px;
  gap: 10px;
}

.ongoing-clock-container {
  padding: 0px 25px 25px;
}

.uk-countdown-number {
  font-size: 1rem;
}

.ongoing-clock-row {
  font-size: 1rem;
  display: flex;
  padding: 10px;
  gap: 25px;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #e5e5e5;
}

.record-card-container {
  padding: 25px;
}

.records-container {
  display: flex;
  align-items: center;
  flex-direction: column;
}
</style>
